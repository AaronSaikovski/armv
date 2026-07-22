// Port of pkg/utils/jsonutils.go PrettyJsonString, i.e. Go's
// encoding/json.Indent with prefix "" and indent "    " (4 spaces).
//
// This is a token-level re-indenter, NOT a serde round-trip: string bytes
// (including escape sequences) and number bytes are copied verbatim, key
// order is preserved, and empty objects/arrays stay compact ("{}"/"[]").
// Invalid JSON (including trailing garbage) returns Err; the caller falls
// back to the raw string, matching the Go error path.

const INDENT: &str = "    ";

/// Returns the 4-space-indented form of `input`, or an error if `input`
/// is not a single valid JSON value.
pub fn pretty_json_string(input: &str) -> anyhow::Result<String> {
    let mut p = Indenter {
        src: input.as_bytes(),
        pos: 0,
        out: String::with_capacity(input.len() * 2),
    };
    p.skip_ws();
    p.value(0)?;
    p.skip_ws();
    if p.pos != p.src.len() {
        anyhow::bail!("invalid character after top-level value");
    }
    Ok(p.out)
}

struct Indenter<'a> {
    src: &'a [u8],
    pos: usize,
    out: String,
}

impl<'a> Indenter<'a> {
    fn skip_ws(&mut self) {
        while let Some(&c) = self.src.get(self.pos) {
            if matches!(c, b' ' | b'\t' | b'\n' | b'\r') {
                self.pos += 1;
            } else {
                break;
            }
        }
    }

    fn peek(&self) -> Option<u8> {
        self.src.get(self.pos).copied()
    }

    fn newline(&mut self, depth: usize) {
        self.out.push('\n');
        for _ in 0..depth {
            self.out.push_str(INDENT);
        }
    }

    fn value(&mut self, depth: usize) -> anyhow::Result<()> {
        match self.peek() {
            Some(b'{') => self.object(depth),
            Some(b'[') => self.array(depth),
            Some(b'"') => self.string(),
            Some(b't') => self.literal("true"),
            Some(b'f') => self.literal("false"),
            Some(b'n') => self.literal("null"),
            Some(c) if c == b'-' || c.is_ascii_digit() => self.number(),
            Some(_) => anyhow::bail!("invalid character looking for beginning of value"),
            None => anyhow::bail!("unexpected end of JSON input"),
        }
    }

    fn object(&mut self, depth: usize) -> anyhow::Result<()> {
        self.pos += 1; // '{'
        self.out.push('{');
        self.skip_ws();
        if self.peek() == Some(b'}') {
            self.pos += 1;
            self.out.push('}');
            return Ok(());
        }
        loop {
            self.newline(depth + 1);
            if self.peek() != Some(b'"') {
                anyhow::bail!("invalid character looking for beginning of object key string");
            }
            self.string()?;
            self.skip_ws();
            if self.peek() != Some(b':') {
                anyhow::bail!("invalid character after object key");
            }
            self.pos += 1;
            self.out.push_str(": ");
            self.skip_ws();
            self.value(depth + 1)?;
            self.skip_ws();
            match self.peek() {
                Some(b',') => {
                    self.pos += 1;
                    self.out.push(',');
                    self.skip_ws();
                }
                Some(b'}') => {
                    self.pos += 1;
                    self.newline(depth);
                    self.out.push('}');
                    return Ok(());
                }
                _ => anyhow::bail!("invalid character after object key:value pair"),
            }
        }
    }

    fn array(&mut self, depth: usize) -> anyhow::Result<()> {
        self.pos += 1; // '['
        self.out.push('[');
        self.skip_ws();
        if self.peek() == Some(b']') {
            self.pos += 1;
            self.out.push(']');
            return Ok(());
        }
        loop {
            self.newline(depth + 1);
            self.value(depth + 1)?;
            self.skip_ws();
            match self.peek() {
                Some(b',') => {
                    self.pos += 1;
                    self.out.push(',');
                    self.skip_ws();
                }
                Some(b']') => {
                    self.pos += 1;
                    self.newline(depth);
                    self.out.push(']');
                    return Ok(());
                }
                _ => anyhow::bail!("invalid character after array element"),
            }
        }
    }

    /// Copies a string token verbatim, validating escapes and control chars
    /// exactly as strictly as Go's scanner does.
    fn string(&mut self) -> anyhow::Result<()> {
        let start = self.pos;
        self.pos += 1; // opening '"'
        loop {
            match self.peek() {
                None => anyhow::bail!("unexpected end of JSON input in string"),
                Some(b'"') => {
                    self.pos += 1;
                    self.out
                        .push_str(std::str::from_utf8(&self.src[start..self.pos])?);
                    return Ok(());
                }
                Some(b'\\') => {
                    self.pos += 1;
                    match self.peek() {
                        Some(b'"' | b'\\' | b'/' | b'b' | b'f' | b'n' | b'r' | b't') => {
                            self.pos += 1;
                        }
                        Some(b'u') => {
                            self.pos += 1;
                            for _ in 0..4 {
                                match self.peek() {
                                    Some(h) if h.is_ascii_hexdigit() => self.pos += 1,
                                    _ => anyhow::bail!("invalid \\u escape in string"),
                                }
                            }
                        }
                        _ => anyhow::bail!("invalid escape in string"),
                    }
                }
                Some(c) if c < 0x20 => {
                    anyhow::bail!("invalid control character in string")
                }
                Some(_) => self.pos += 1,
            }
        }
    }

    /// Copies a number token verbatim, validating the JSON number grammar
    /// (-? (0 | [1-9][0-9]*) frac? exp?).
    fn number(&mut self) -> anyhow::Result<()> {
        let start = self.pos;
        if self.peek() == Some(b'-') {
            self.pos += 1;
        }
        match self.peek() {
            Some(b'0') => self.pos += 1,
            Some(c) if c.is_ascii_digit() => {
                while matches!(self.peek(), Some(d) if d.is_ascii_digit()) {
                    self.pos += 1;
                }
            }
            _ => anyhow::bail!("invalid number"),
        }
        if self.peek() == Some(b'.') {
            self.pos += 1;
            if !matches!(self.peek(), Some(d) if d.is_ascii_digit()) {
                anyhow::bail!("invalid number fraction");
            }
            while matches!(self.peek(), Some(d) if d.is_ascii_digit()) {
                self.pos += 1;
            }
        }
        if matches!(self.peek(), Some(b'e' | b'E')) {
            self.pos += 1;
            if matches!(self.peek(), Some(b'+' | b'-')) {
                self.pos += 1;
            }
            if !matches!(self.peek(), Some(d) if d.is_ascii_digit()) {
                anyhow::bail!("invalid number exponent");
            }
            while matches!(self.peek(), Some(d) if d.is_ascii_digit()) {
                self.pos += 1;
            }
        }
        self.out
            .push_str(std::str::from_utf8(&self.src[start..self.pos])?);
        Ok(())
    }

    fn literal(&mut self, lit: &str) -> anyhow::Result<()> {
        if self.src[self.pos..].starts_with(lit.as_bytes()) {
            self.pos += lit.len();
            self.out.push_str(lit);
            Ok(())
        } else {
            anyhow::bail!("invalid literal")
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn compact_object_is_indented() {
        let got = pretty_json_string(r#"{"a":1,"b":"x"}"#).unwrap();
        assert_eq!(got, "{\n    \"a\": 1,\n    \"b\": \"x\"\n}");
    }

    #[test]
    fn nested_structures() {
        let got = pretty_json_string(r#"{"a":{"b":[1,2]}}"#).unwrap();
        assert_eq!(
            got,
            "{\n    \"a\": {\n        \"b\": [\n            1,\n            2\n        ]\n    }\n}"
        );
    }

    #[test]
    fn empty_object_and_array_stay_compact() {
        assert_eq!(pretty_json_string("{}").unwrap(), "{}");
        assert_eq!(pretty_json_string("[]").unwrap(), "[]");
        assert_eq!(
            pretty_json_string(r#"{"a":{},"b":[]}"#).unwrap(),
            "{\n    \"a\": {},\n    \"b\": []\n}"
        );
    }

    #[test]
    fn already_pretty_is_renormalised() {
        let got = pretty_json_string("{\n  \"a\": 1\n}").unwrap();
        assert_eq!(got, "{\n    \"a\": 1\n}");
    }

    #[test]
    fn key_order_preserved() {
        let got = pretty_json_string(r#"{"z":1,"a":2,"m":3}"#).unwrap();
        assert_eq!(got, "{\n    \"z\": 1,\n    \"a\": 2,\n    \"m\": 3\n}");
    }

    #[test]
    fn escapes_and_numbers_verbatim() {
        let got = pretty_json_string(r#"{"s":"aA\n\\|","n":1.50,"e":1e5,"z":-0}"#).unwrap();
        assert_eq!(
            got,
            "{\n    \"s\": \"aA\\n\\\\|\",\n    \"n\": 1.50,\n    \"e\": 1e5,\n    \"z\": -0\n}"
        );
    }

    #[test]
    fn scalars_at_top_level() {
        assert_eq!(pretty_json_string("5").unwrap(), "5");
        assert_eq!(pretty_json_string("\"x\"").unwrap(), "\"x\"");
        assert_eq!(pretty_json_string("true").unwrap(), "true");
        assert_eq!(pretty_json_string("null").unwrap(), "null");
    }

    #[test]
    fn surrounding_whitespace_allowed() {
        assert_eq!(pretty_json_string("  {\"a\":1}  \n").unwrap(), "{\n    \"a\": 1\n}");
    }

    #[test]
    fn invalid_inputs_error() {
        for bad in [
            "", "{", "[1,]", "{\"a\":}", "tru", "1 2", "{\"a\":1}}", "<html>", "{'a':1}",
            "{\"a\" 1}", "01", "-", "1.", "1e", "\"unterminated", "\"bad\\qescape\"",
        ] {
            assert!(pretty_json_string(bad).is_err(), "expected error for {bad:?}");
        }
    }

    #[test]
    fn output_reparses_as_json() {
        let src = r#"{"error":{"code":"X","details":[{"target":"/a/b","message":"m|m"}]}}"#;
        let got = pretty_json_string(src).unwrap();
        let a: serde_json::Value = serde_json::from_str(src).unwrap();
        let b: serde_json::Value = serde_json::from_str(&got).unwrap();
        assert_eq!(a, b);
    }
}
