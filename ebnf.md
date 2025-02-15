# ebnf

## Lexical Elements

```ebnf
(* 
  Lexical Elements 
  ---------------- 
  Note: The productions for Letter and Digit are given in an abbreviated form. 
*)

Letter         = "A" | "B" | "C" | "D" | "E" | "F" | "G" | "H" | "I" | "J" |
                 "K" | "L" | "M" | "N" | "O" | "P" | "Q" | "R" | "S" | "T" |
                 "U" | "V" | "W" | "X" | "Y" | "Z" |
                 "a" | "b" | "c" | "d" | "e" | "f" | "g" | "h" | "i" | "j" |
                 "k" | "l" | "m" | "n" | "o" | "p" | "q" | "r" | "s" | "t" |
                 "u" | "v" | "w" | "x" | "y" | "z" ;

Digit          = "0" | "1" | "2" | "3" | "4" | "5" | "6" | "7" | "8" | "9" ;

HexDigit       = Digit | "a" | "b" | "c" | "d" | "e" | "f" | "A" | "B" | "C" | "D" | "E" | "F" ;

Name           = Letter { Letter | Digit | "_" } ;
Identifier     = Name { "." Name } ;

WS             = { " " | "\t" } ;
NewLine        = "\n" | "\r\n" ;

AnyCharacter   = ? any character ? ;
```

## Markdown Files

```ebnf
(* 
  Markdown Files 
  -------------- 
  A Markdown file consists of a header (with a package declaration) followed by a body. 
*)

MarkdownFile   = Header Body ;

Header         = "---" NewLine
                 PackageDeclaration NewLine
                 "---" NewLine ;
PackageDeclaration = "reqmd.package:" WS PackageID ;
PackageID      = Identifier ;

Body           = { MarkdownElement } ;
MarkdownElement = RequirementSite | CoverageFootnote | PlainText ;

PlainText      = { AnyCharacter } ;

(* 
  Requirement Sites in Markdown 
  ------------------------------ 
  A requirement site is written in the text as a backtick‐quoted requirement ID. 
  An annotated requirement site includes coverage status and a footnote reference. 
*)

RequirementSite = RequirementSiteLabel [ CoverageStatus CoverageFootnoteReference ] [ CoverageStatusEmoji ] ;
CoverageStatusWord = "covered" | "uncvrd" ;
CoverageStatusEmoji = "✅" | "❓" ;
RequirementSiteLabel = "`" RequirementSiteID "`" ;
RequirementSiteID = "~" RequirementName "~" ;
RequirementName = Identifier ;
RequirementID = PackageID "/" RequirementName ;

CoverageFootnoteReference = "[^" RequirementSiteID "]" ;

(* 
  Coverage Footnotes in Markdown 
  ------------------------------ 
  A coverage footnote links a requirement (via its requirement site ID) to a hint and 
  optional comma-separated coverers. 
*)

CoverageFootnote = "[^" RequirementSiteID "]:" CoverageFootnoteHint [ CovererList ] ;
CoverageFootnoteHint = "`[" "~" PackageID "/" RequirementName "~impl]`" ;
CovererList    = Coverer { "," WS Coverer } ;
Coverer        = "[" CoverageLabel "]" "(" CoverageURL ")" ;
CoverageLabel  = FilePath ":" "line" Digit { Digit } ":" CoverageType ;

(* 
  Coverage URL 
  ------------ 
  A coverage URL points to a specific location in a source code repository. 
*)

CoverageURL    = FileURL [ "?plain=1" ] "#" CoverageArea ;
FileURL        = GitHubURL | GitLabURL ;
GitHubURL      = GitHubBaseURL "/blob/" CommitHash "/" FilePath ;
GitHubBaseURL  = "https://github.com/" Owner "/" Repository ;
GitLabURL      = GitLabBaseURL "/-/blob/" CommitHash "/" FilePath ;
GitLabBaseURL  = "https://gitlab.com/" Owner "/" Repository ;
Owner          = Identifier ;
Repository     = Identifier ;
CommitHash     = HexDigit { HexDigit } ;
FilePath       = { AnyCharacter - ("?" | "#") } ;
CoverageArea   = "L" Digit { Digit } ;
```

## Source Files

```ebnf

(* 
  Source Files 
  ------------ 
  A source file is a text file that may contain one or more CoverageTags. 
  A CoverageTag is written as a bracketed expression which links a requirement (by its package and name) 
  to a coverage type. 
  
  For example: 
      [~server.api.v2/Post.handler~test] 
  Here, PackageID is "server.api.v2", RequirementName is "Post.handler", and CoverageType is "test". 
*)

SourceFile   = { SourceElement } ;
SourceElement = CoverageTag | PlainText ;
CoverageTag  = "[" "~" PackageID "/" RequirementName "~" CoverageType "]";
```
