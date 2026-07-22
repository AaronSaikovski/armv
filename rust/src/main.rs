// Port of cmd/armv/main.go: build metadata, signal wiring, error surfacing.
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

        armv::app::run(
            cancel,
            armv::app::Config {
                version: full_version(),
                args,
            },
        )
        .await
    });

    match result {
        Ok(()) => ExitCode::SUCCESS,
        Err(err) => {
            eprintln!("Error: {err:#}");
            ExitCode::FAILURE
        }
    }
}
