# Dinkur

Task and time tracking utility.

## Install

```console
$ go install --tags fts5 github.com/dinkur/dinkur@latest
```

> The `--tags fts5` flags adds [Sqlite FTS5](https://www.sqlite.org/fts5.html)
> support, which is used for better and more performant full-text search.

For you CLI-power users, we recommend aliasing it to `ur`:

<table><thead><tr>
<th>Bash/Zsh/Fish</th>
<th>Powershell</th>
</tr></thead><tbody><tr><td>

```sh
alias ur=dinkur
```

</td><td>

```powershell
Set-Alias -Name ur -Value dinkur
```

</td></tr></tbody></table>

## Development

Requires Go v1.17 (or higher)

```sh
go run --tags fts5 .
```
