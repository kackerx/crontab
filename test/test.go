package main

import (
	"encoding/json"
	"fmt"
)

type Test struct {
	Name string
}

func main() {
	t := Test{Name: "kacker"}
	fmt.Println(t.Name)
	s := `
{
	"name": "kacker", 
	"age": "24"
}
`
	m := make(map[string]string)
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		fmt.Println(err)
	}
	fmt.Println(m)
	test("kacker", 24)
}

// test 颜色
func test(name string, age int) string {
	return name
}
