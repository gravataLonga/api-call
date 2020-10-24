# API CALL - UTP  

![Test](https://github.com/gravataLonga/api-call/workflows/Test/badge.svg?branch=master)  

## How to use  

### Create configuration

```
apiCall := apicall.New(
    WithMethod("GET"),
    WithUrl("/"),
)
```

> Tip: You can create your own method for configuration, you only need to implement Option type.  

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

