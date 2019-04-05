# sqljoiner

> combine sql files into a single schema.

## Overview
sqljoiner combines all of the postgres sql files in a directory
into a single sql file that can be used to genenerate the full schema.
They are combined in order based on their dependencies to other tables, views, etc.

This along with a tool like [migra](https://github.com/djrobstep/migra) allows you to have your current database schema
broken up into individual files and have migrations generated for you when you change the current schema.

## Example
`sqljoiner -dir example/`
