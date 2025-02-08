# ebnf

```ebnf
MarkdownFile = MarkdownHeader MarkdownBody
```

```ebnf
(* Identifiers and their basic parts follow the provided definitions. *)
Identifier        = Letter { Letter | Digit | "_" }

(* A package name is a sequence of one or more identifiers separated by dots. *)
PackageName       = Identifier { "." Identifier }

(* A Markdown file consists of a header and a body. *)
MarkdownHeader  = HeaderDelimiter NewLine HeaderBody HeaderDelimiter NewLine

(* The header delimiter is three dashes. *)
HeaderDelimiter = "---"

(* The body is zero or more header lines. *)
HeaderBody      = { HeaderLine NewLine }

(* A header line is either a field line (such as the reqmd.package field)
   or any other line of text. *)
HeaderLine      = FieldLine | OtherLine

(* A header line is either a field line (such as the reqmd.package field)
   or any other line of text. *)
HeaderLine      = FieldLine | OtherLine

(* The reqmd.package field line uses a fixed key, a colon, optional whitespace,
   and a package field value (the package name enclosed in "<" and ">"). *)
FieldLine       = "reqmd.package:" PackageName

```

```ebnf
(* A MarkdownBody is a sequence of text elements. *)
MarkdownBody      = { TextElement }

(* A TextElement is either a RequirementID, a CoveringFootnote on its own, or any other text. *)
TextElement       = RequirementID | CoveringFootnote | OtherText

(* A RequirementID is an inline requirement element that may optionally be annotated
   with coverage information. *)
RequirementID      = InlineReqID [ CoverageAnnotation ]

(* The inline part of a RequirementID is written as a code element that begins with 
   a tilde, contains a requirement name, and ends with a tilde. *)
InlineReqID        = "`" "~" Identifier "~" "`"

(* A covered RequirementID is marked by the literal "coverers" immediately following the inline code, followed by a CoveringFootnote. *)
CoverageAnnotation = "coverers" FootnoteReference

FootnoteReference = FootnoteLabel

FootnoteLabel = "[^" CoverersID "]"

CoverersID = "coverers" NUMBER

(* A CoveringFootnote is a footnote marker followed by a CoverageTag and a FileCoverageList. *)
CoveringFootnote  = FootnoteMarker "`" CoverageTag "`" ":" FileCoverageList

FootnoteMarker    = FootnoteLabel ":"
OtherText          ::= { AnyChar } 
```

```ebnf


(* A requirement identifier is a package name, a literal slash, and a requirement name. *)
RequirementIdentifier = PackageName "/" RequirementName

RequirementName   = Identifier
```

Coverage

```ebnf


CoverageTag       = "[~" RequirementIdentifier "~" CoverageType "]"

CoverageType      = Identifier

(* A list of file coverage entries is one or more FileCoverage items separated by commas. *)
FileCoverageList  = FileCoverage { "," FileCoverage }

(* Each file coverage entry consists of a file reference immediately followed by a coverage URL. *)
FileCoverage      = CoverageLabel CoverageURL

(* The file reference is enclosed in square brackets and has the form: 
   file path, colon, line specification, colon, a coverage type. *)
CoverageLabel     = "[" FilePath ":" Line ":" CoverageType "]"

(* A file path is given as one or more path components (identifiers) separated by a slash. *)
FilePath          = Identifier { "/" Identifier }

(* A line specification begins with the literal "line" followed by one or more digits. *)
Line              = "line" Digit { Digit }

(* The coverage URL is provided in parentheses. (This production may be adapted
   to a fuller URL grammar as needed.) *)
CoverageURL       = "(" URL ")"
```
