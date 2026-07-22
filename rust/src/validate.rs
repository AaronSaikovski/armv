// Port of pkg/utils/validateinput.go. The Go regex is
// ^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$
// implemented here as a structural check to avoid a regex dependency.
// Anchors are load-bearing: braces, whitespace, embedded newlines, and any
// leading/trailing garbage must all be rejected.

/// Reports whether `subscription_id` is a well-formed bare (unbraced) UUID.
pub fn check_valid_subscription_id(subscription_id: &str) -> bool {
    let b = subscription_id.as_bytes();
    if b.len() != 36 {
        return false;
    }
    for (i, &c) in b.iter().enumerate() {
        match i {
            8 | 13 | 18 | 23 => {
                if c != b'-' {
                    return false;
                }
            }
            _ => {
                if !c.is_ascii_hexdigit() {
                    return false;
                }
            }
        }
    }
    true
}

#[cfg(test)]
mod tests {
    use super::check_valid_subscription_id;

    const VALID: &str = "12345678-1234-1234-1234-123456789012";

    #[test]
    fn valid_uuids() {
        assert!(check_valid_subscription_id(VALID));
        assert!(check_valid_subscription_id("abcdefab-cdef-abcd-efab-cdefabcdefab"));
        assert!(check_valid_subscription_id("ABCDEFAB-CDEF-ABCD-EFAB-CDEFABCDEFAB"));
        assert!(check_valid_subscription_id("aBcDeFaB-cDeF-1234-EfAb-cDeFaBcDeF12"));
        assert!(check_valid_subscription_id("00000000-0000-0000-0000-000000000000"));
    }

    #[test]
    fn invalid_uuids() {
        for bad in [
            "",
            "12345678-1234-1234-1234-12345678901",   // too short
            "12345678-1234-1234-1234-1234567890123", // too long
            "123456781234123412341234567890120000",  // no hyphens
            "12345678-1234-1234-1234-12345678901g",  // non-hex
            "{12345678-1234-1234-1234-123456789012}", // braced
            "{12345678-1234-1234-1234-123456789012", // unbalanced brace
            " 12345678-1234-1234-1234-123456789012", // leading space
            "12345678-1234-1234-1234-123456789012 ", // trailing space
            "12345678-1234-1234-1234-123456789012\n", // trailing newline
            "12345678-1234-1234-1234-123456789012\nevil", // embedded newline
            "12345678_1234_1234_1234_123456789012", // wrong separators
            "1234567-81234-1234-1234-123456789012", // hyphens misplaced
        ] {
            assert!(!check_valid_subscription_id(bad), "expected invalid: {bad:?}");
        }
    }
}
