# Crawlee

![Crawlee](crawleev2.png)

Simple web scraper in Go

## Usage

1. Add URLs to `urls.txt`
2. Run: `go run main.go`
3. Check results in `raw-results/` folder

## Example

```
# urls.txt
https://www.example.com/
https://www.news-site.com/
```

Results saved as:

- `raw-results/example.txt`
- `raw-results/news-site.txt`

## Todo

- [ ] Add filtering already stored results
