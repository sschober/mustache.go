package mustache

import (
    "container/vector"
    "os"
    "path"
    "strings"
    "testing"
    //    "fmt"
)

type Test struct {
    tmpl     string
    context  interface{}
    expected string
}

type Data struct {
    a   bool
    b   string
}

type User struct {
    Name string
    Id   int64
}

func (u User) AddOne() int64 {
    res := u.Id + 1
    return res
}

func (u *User) AddOneStar() int64 { return u.AddOne() }

func (u User) ConcatWorld() string {
    res := u.Name + " World"
    return res
}

func makeVector(n int) *vector.Vector {
    v := new(vector.Vector)
    for i := 0; i < n; i++ {
        v.Push(&User{"Mike", 1})
    }
    return v
}

var tests = []Test{
    Test{`hello {{name}}`, map[string]string{"name": "world"}, "hello world"},
    Test{`{{a}}{{b}}{{c}}{{d}}`, map[string]string{"a": "a", "b": "b", "c": "c", "d": "d"}, "abcd"},
    Test{`0{{a}}1{{b}}23{{c}}456{{d}}89`, map[string]string{"a": "a", "b": "b", "c": "c", "d": "d"}, "0a1b23c456d89"},
    Test{`hello {{! comment }}world`, map[string]string{}, "hello world"},
    Test{`{{ a }}{{=<% %>=}}<%b %><%={{ }}=%>{{ c }}`, map[string]string{"a": "a", "b": "b", "c": "c"}, "abc"},
    Test{`{{ a }}{{= <% %> =}}<%b %><%= {{ }}=%>{{c}}`, map[string]string{"a": "a", "b": "b", "c": "c"}, "abc"},

    //section tests
    Test{`{{#a}}{{b}}{{/a}}`, Data{true, "hello"}, "hello"},
    Test{`{{#a}}{{b}}{{/a}}`, Data{false, "hello"}, ""},
    Test{`{{a}}{{#b}}{{b}}{{/b}}{{c}}`, map[string]string{"a": "a", "b": "b", "c": "c"}, "abc"},
    Test{`{{#a}}{{b}}{{/a}}`, struct {
        a []struct {
            b string
        }
    }{[]struct {
        b string
    }{struct{ b string }{"a"}, struct{ b string }{"b"}, struct{ b string }{"c"}}},
        "abc",
    },
    Test{`{{#a}}{{b}}{{/a}}`, struct{ a []map[string]string }{[]map[string]string{map[string]string{"b": "a"}, map[string]string{"b": "b"}, map[string]string{"b": "c"}}}, "abc"},
    Test{`{{#users}}{{Name}}{{/users}}`, map[string]interface{}{"users": []User{User{"Mike", 1}}}, "Mike"},
    Test{`{{#users}}{{Name}}{{/users}}`, map[string]interface{}{"users": []*User{&User{"Mike", 1}}}, "Mike"},
    Test{`{{#users}}{{Name}}{{/users}}`, map[string]interface{}{"users": vector.Vector([]interface{}{&User{"Mike", 12}})}, "Mike"},
    Test{`{{#users}}{{Name}}{{/users}}`, map[string]interface{}{"users": makeVector(1)}, "Mike"},
    Test{`{{Name}}`, User{"Mike", 1}, "Mike"},
    Test{`{{Name}}`, &User{"Mike", 1}, "Mike"},
    Test{`{{AddOne}}`, User{"Mike", 1}, "2"},
    Test{`{{AddOne}}`, &User{"Mike", 1}, "2"},
    // The next test doesn't work due to reflection issues (see mustache.go:func lookup)
    //Test{`{{AddOneStar}}`, User{"Mike", 1}, "2"},
    Test{`{{AddOneStar}}`, &User{"Mike", 1}, "2"},
    Test{`{{ConcatWorld}}`, User{"Mike", 1}, "Mike World"},
    Test{"{{#users}}\n{{Name}}\n{{/users}}", map[string]interface{}{"users": makeVector(2)}, "Mike\nMike\n"},
    Test{"{{#users}}\r\n{{Name}}\r\n{{/users}}", map[string]interface{}{"users": makeVector(2)}, "Mike\r\nMike\r\n"},
}

func TestBasic(t *testing.T) {
    for _, test := range tests {
        output, err := Render(test.tmpl, test.context)
        if err != nil {
            t.Fatalf("%q got error %q", test.tmpl, err.String())
        } else if output != test.expected {
            t.Fatalf("%q expected %q got %q", test.tmpl, test.expected, output)
        }
    }
}

func TestFile(t *testing.T) {
    filename := path.Join(path.Join(os.Getenv("PWD"), "tests"), "test1.mustache")
    expected := "hello world"
    output, err := RenderFile(filename, map[string]string{"name": "world"})
    if err != nil {
        t.Fatalf("Error in test1.mustache", err.String())
    } else if output != expected {
        t.Fatalf("testfile expected %q got %q", expected, output)
    }
}

func TestPartial(t *testing.T) {
    filename := path.Join(path.Join(os.Getenv("PWD"), "tests"), "test2.mustache")
    expected := "hello world"
    output, err := RenderFile(filename, map[string]string{"name": "world"})
    if err != nil {
        t.Fatalf("Error in test2.mustache", err.String())
    } else if output != expected {
        t.Fatalf("testpartial expected %q got %q", expected, output)
    }
}

var malformed = []Test{
    Test{`{{#a}}{{}}{{/a}}`, Data{true, "hello"}, "empty tag"},
    Test{`{{}}`, nil, "empty tag"},
    Test{`{{}`, nil, "unmatched open tag"},
    Test{`{{`, nil, "unmatched open tag"},
}

func TestMalformed(t *testing.T) {
    for _, test := range malformed {
        _, err := Render(test.tmpl, test.context)
        if err == nil {
            t.Fatalf("%q expected error, got none", test.tmpl)
        } else if strings.Index(err.String(), test.expected) == -1 {
            t.Fatalf("%q expected %q in error %q", test.tmpl, test.expected, err.String())
        }
    }
}
