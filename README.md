# API CALL - [Creative Code Solution](https://creativecodesolutions.pt)     

![Test](https://github.com/gravataLonga/api-call/workflows/Test/badge.svg?branch=master)
[![Coverage Status](https://coveralls.io/repos/github/gravataLonga/api-call/badge.svg?branch=master)](https://coveralls.io/github/gravataLonga/api-call?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/gravataLonga/api-call)](https://goreportcard.com/report/github.com/gravataLonga/api-call)

Read [documentation here](https://pkg.go.dev/mod/github.com/gravataLonga/api-call).

## How to use  

### Create configuration

```
apiCall := apicall.New(
    WithBaseUrl("https://www.google.pt"),
)
```

> Tip: You can create your own method for configuration, you only need to implement Option type.  

### Handler Response  

```
response, err := apiCall.Send("GET", "/users/list", nil)  
if err != nil {
    panic("An error happen")
}

if !response.IsOk() {
    panic("Unable to success response")
}

fmt.Println(response) // response is apicall.BaseStandard struct
```

### Bind your own structure to response  

```
response, err := apiCall.Send("POST", "/get-users", nil)
if err != nil {
    panic("An error happen")
}

if !response.IsOk() {
    panic("Unable to success response")
}

type User struct {
    Name string
    Email string
}

var users []User
err = response.GetItems(&users)

for index, user := range users {
    fmt.Println(user.Name)
}
```

