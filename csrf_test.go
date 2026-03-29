package main
import (
    "fmt"
    "net/url"
)
func main() {
    u, _ := url.Parse("https://url.taverns.red")
    fmt.Println(u.Host)
}
