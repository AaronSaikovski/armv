// Port of cmd/armv/main.go: build metadata, signal wiring, error surfacing.
// `unsafe` is forbidden package-wide via `[lints]` in Cargo.toml.
use std::process::ExitCode;

use tokio_util::sync::CancellationToken;

// Injected by build.rs (env override -> git -> Go zero-values dev/none/unknown),
// mirroring the Go -ldflags -X main.version/main.commit/main.date mechanism.
const VERSION: &str = env!("ARMV_VERSION");
const COMMIT: &str = env!("ARMV_COMMIT");
const DATE: &str = env!("ARMV_DATE");

/// The full version string shown by --version.
fn full_version() -> String {
    format!("{VERSION} (commit {COMMIT}, built {DATE})")
}

/// Installs a `tracing` subscriber that writes verbose diagnostics to both
/// stderr and a timestamped `armv-debug-*.log` file in the output directory,
/// but only when `--debug` is set — so normal runs stay quiet and byte-exact.
/// `RUST_LOG` overrides the default `armv=debug` filter when present.
fn init_logging(debug: bool, output_path: &str) {
    if !debug {
        return;
    }
    use tracing_subscriber::{fmt, prelude::*, EnvFilter};

    let filter =
        EnvFilter::try_from_default_env().unwrap_or_else(|_| EnvFilter::new("armv=debug"));

    let stderr_layer = fmt::layer()
        .with_target(false)
        .with_writer(std::io::stderr);

    // Best-effort file sink; if the log file can't be created we still log to
    // stderr. The file layer has no ANSI colour codes.
    let file_layer = open_debug_log(output_path).map(|file| {
        fmt::layer()
            .with_ansi(false)
            .with_target(false)
            .with_writer(file)
    });

    let _ = tracing_subscriber::registry()
        .with(filter)
        .with(stderr_layer)
        .with(file_layer)
        .try_init();
}

/// Creates the output directory and opens a timestamped debug-log file in it.
/// Returns a shared, thread-safe writer, or `None` on any I/O error.
fn open_debug_log(output_path: &str) -> Option<SharedFile> {
    let file_name = chrono::Local::now()
        .format("armv-debug-%Y-%m-%d-%H-%M-%S.log")
        .to_string();
    std::fs::create_dir_all(output_path).ok()?;
    let path = std::path::Path::new(output_path).join(file_name);
    let file = std::fs::File::create(&path).ok()?;
    eprintln!("Debug log: {}", path.display());
    Some(SharedFile(std::sync::Arc::new(std::sync::Mutex::new(file))))
}

/// A thread-safe `MakeWriter` over a single file handle (the pipeline logs
/// from multiple runtime threads).
#[derive(Clone)]
struct SharedFile(std::sync::Arc<std::sync::Mutex<std::fs::File>>);

impl<'a> tracing_subscriber::fmt::MakeWriter<'a> for SharedFile {
    type Writer = SharedFileHandle;
    fn make_writer(&'a self) -> Self::Writer {
        SharedFileHandle(self.0.clone())
    }
}

struct SharedFileHandle(std::sync::Arc<std::sync::Mutex<std::fs::File>>);

impl std::io::Write for SharedFileHandle {
    fn write(&mut self, buf: &[u8]) -> std::io::Result<usize> {
        // Recover from a poisoned lock rather than panicking: a logging sink
        // must never be able to take down the process.
        self.0.lock().unwrap_or_else(|e| e.into_inner()).write(buf)
    }
    fn flush(&mut self) -> std::io::Result<()> {
        self.0.lock().unwrap_or_else(|e| e.into_inner()).flush()
    }
}

fn main() -> ExitCode {
    let argv: Vec<String> = std::env::args().skip(1).collect();
    let args = match armv::cli::parse(&argv, &full_version()) {
        Ok(Some(args)) => args,
        Ok(None) => return ExitCode::SUCCESS,
        Err(err) => {
            eprintln!("Error: {err:#}");
            return ExitCode::FAILURE;
        }
    };

    init_logging(args.debug, &args.output_path);

    let runtime = tokio::runtime::Runtime::new().expect("failed to build tokio runtime");
    let result = runtime.block_on(async {
        // Cancel on Ctrl-C / SIGTERM so the poll loop unwinds cleanly
        // (finishes the progress bar and returns a cancellation error),
        // mirroring Go's signal.NotifyContext.
        let cancel = CancellationToken::new();
        let signal_cancel = cancel.clone();
        tokio::spawn(async move {
            #[cfg(unix)]
            {
                use tokio::signal::unix::{signal, SignalKind};
                let mut term = signal(SignalKind::terminate()).expect("SIGTERM handler");
                tokio::select! {
                    _ = tokio::signal::ctrl_c() => {}
                    _ = term.recv() => {}
                }
            }
            #[cfg(not(unix))]
            {
                let _ = tokio::signal::ctrl_c().await;
            }
            signal_cancel.cancel();
        });

        armv::app::run(cancel, armv::app::Config { args }).await
    });

    match result {
        Ok(()) => ExitCode::SUCCESS,
        Err(err) => {
            eprintln!("Error: {err:#}");
            ExitCode::FAILURE
        }
    }
}
