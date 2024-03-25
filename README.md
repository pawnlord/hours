# hours
Go script to count how long you've worked on a project

## usage
Run `go run /path/to/hours.go` to start the script. It will run in the background, with a CLI.

### .hoursignore.json
If there are files you don't want to be scraped for modifications, you can create a `.hoursignore.json` in the root directory. The format is as follows:
```json
[
    {
        "Pattern": "./go.mod",
        "IsNeg": false
    }, ...
]
```
- **Pattern** follows the same format as the `pattern` argument to the go `path.Match` function
- **IsNeg** denotes if this pattern should be an exception to ignored files. If it's false, the pattern is ignored. If it's true, a file that matches the pattern will be checked.