Search the content of files, within specific directories, using a CLI.

Currently only supports prefix, case insensitive, searching.

## Usage
```
sol pathToScan [-EE space delimited list]
-EE: excluded extensions, files with these extensions will not be searched; for example -EE exe sql

During execution: [-B int] [-A int] search[*]
-B: print num lines of leading context before matching line.
-A: print num lines of trailing context after matching line.
*: do a prefix search, rather than a whole word search.

Note flags can be placed anywhere, e.g. this is valid: [-B int] search [-A int]
```

## Config
On first execution, a `~/.sol/.solconfig` file will be created.

This file holds the default excluded directories, and extensions (these are excluded from all search results).
