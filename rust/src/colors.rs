// ANSI color helpers matching github.com/logrusorgru/aurora v2.0.3 output
// byte-for-byte. Derived from aurora's value.go/color.go:
//   Green(x)       -> "\x1b[32m" + x + "\x1b[0m"
//   Yellow(x)      -> "\x1b[33m" + x + "\x1b[0m"
//   Red(x)         -> "\x1b[31m" + x + "\x1b[0m"
//   Bold(Green(x)) -> "\x1b[1;32m" + x + "\x1b[0m"
//   Bold(Red(x))   -> "\x1b[1;31m" + x + "\x1b[0m"
// aurora.Sprintf(Green(fmt), Green(arg)) renders the arg with a "tail" color
// re-opening the outer color after it: args become "\x1b[0;32m" + arg +
// "\x1b[0;32m" (note: no reset between arg and tail). Colors are emitted
// unconditionally - aurora does no TTY detection, and neither do we.

pub const RESET: &str = "\x1b[0m";

pub fn green(s: &str) -> String {
    format!("\x1b[32m{s}{RESET}")
}

pub fn bold_green(s: &str) -> String {
    format!("\x1b[1;32m{s}{RESET}")
}

pub fn red(s: &str) -> String {
    format!("\x1b[31m{s}{RESET}")
}

pub fn bold_red(s: &str) -> String {
    format!("\x1b[1;31m{s}{RESET}")
}

pub fn yellow(s: &str) -> String {
    format!("\x1b[33m{s}{RESET}")
}

/// Cyan, used for progress/status lines (aurora Cyan = 36). These lines are
/// not present in the Go binary.
pub fn cyan(s: &str) -> String {
    format!("\x1b[36m{s}{RESET}")
}

/// The success banner's status line, replicating
/// `aurora.Sprintf(aurora.Green("*** Response Status OK - %s ***"), aurora.Green(respStatus))`.
pub fn green_status_line(resp_status: &str) -> String {
    format!("\x1b[32m*** Response Status OK - \x1b[0;32m{resp_status}\x1b[0;32m ***{RESET}")
}
