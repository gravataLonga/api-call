# API CALL - UTP  

## How to use  

### Create configuration

```
apiCall := apicall.New("GET", "/users")
```

### Handler Response  

```
response, err := apiCall.Send()
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
response, err := apiCall.Send()
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

