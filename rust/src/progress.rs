// Port of cmd/armv/poller/progressbar.go (schollz/progressbar). Cosmetic
// only - byte parity with schollz is not required. Kept: max 100, the
// " Running Validation..." description, green '=' fill with '>' head,
// "[" / "]" delimiters, and wrap-at-100 driven by tick().

use indicatif::{ProgressBar, ProgressStyle};

use crate::lro::PROGRESS_BAR_MAX;

/// Wrapper around indicatif tracking the Go barCount wrap logic.
pub struct Progress {
    bar: ProgressBar,
    count: u32,
}

impl Progress {
    pub fn new() -> Self {
        let bar = ProgressBar::new(u64::from(PROGRESS_BAR_MAX));
        bar.set_style(
            ProgressStyle::with_template(
                " Running Validation... {percent:>3}% [{bar:40.green}] [{elapsed}]",
            )
            .expect("static progress template")
            .progress_chars("=> "),
        );
        Self { bar, count: 0 }
    }

    /// Advances one tick; resets to zero after reaching 100 (Go increments,
    /// then resets when barCount >= progressBarMax).
    pub fn tick(&mut self) {
        self.bar.inc(1);
        self.count += 1;
        if self.count >= PROGRESS_BAR_MAX {
            self.bar.reset();
            self.count = 0;
        }
    }

    /// Finishes the bar (Go bar.Finish()); called on every exit path.
    pub fn finish(&self) {
        self.bar.finish();
    }
}

impl Default for Progress {
    fn default() -> Self {
        Self::new()
    }
}
