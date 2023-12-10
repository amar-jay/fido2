package main

import (
	"github.com/go-webauthn/webauthn/webauthn"
	fiber "github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
        "github.com/gofiber/fiber/v2/middleware/logger"
        "github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/google/uuid"
)

var webAuthn *webauthn.WebAuthn //not proud of how I initialize this
var appName = "Go Fido2"
var (
    users = make(map[string]*User)
)


type User struct {
    Id string
    Name string
    Credentials []webauthn.Credential
    Session *webauthn.SessionData
}

func(u User) WebAuthnID() []byte {
    return []byte(u.Id)
}

func (u User) WebAuthnCredentials() []webauthn.Credential{
    return u.Credentials
}

func (u User) WebAuthnDisplayName() string {
    return u.Name
}

func (u User) WebAuthnIcon() string {
    return u.Name
}

func (u User)WebAuthnName() string{
    return u.Name
}
	
var _ webauthn.User = User{}

func newUser(name string) (User){
    id := uuid.New().String()
    return User{
        Id: id,
        Name: name,
    }
}

func getUser(name string) (User) {
    user := users[name]
    if user == nil {
        u := newUser(name)
        user = &u
        users[name] = user
    }

    return *user
}

func main() {
    wconfig := &webauthn.Config{
            RPDisplayName: appName, // Display Name for your site
        RPID: "localhost", // Generally the FQDN for your site
        RPOrigins: []string{"https://localhost:8000", "https://localhost:8000"}, // The origin URLs allowed for WebAuthn requests
    }

    var err error
    app := fiber.New()
    app.Use(logger.New())
    app.Use(cors.New())
    webAuthn, err = webauthn.New(wconfig) 
    if err != nil {
        panic(err)
    }
    print("assigned")

    app.Get("/", func(c *fiber.Ctx) error {
        return c.SendString("Hello, World!")
    })

    app.Get("/sign-up/:name", func(c *fiber.Ctx) error {
        err := BeforeRegistration(c) 
        if err != nil {
            return err
        }
        err = AfterRegistration(c)
        if err != nil {
            return err
        }
        return nil
    })

    app.Get("/register/:name", BeforeRegistration)
    app.Get("/authenticate/:name", AfterRegistration)
    app.Get("/handlers", func(c *fiber.Ctx)error{
       return c.JSON(app.GetRoutes())
    })

    //if err := app.ListenTLS(":8000", "./cert.pem", "./key.pem");err != nil {
    if err := app.Listen(":8000");err != nil{
        panic(err)
    }

}

func BeforeRegistration(c *fiber.Ctx) error {
    name := c.Params("name","unknown")
    user := getUser(name)

    options, session, err := webAuthn.BeginRegistration(user)
    if err != nil {
        return fiber.ErrInternalServerError
    }

    user.Session = session
    users[user.Name] = &user // do not understand why this is neccessary

    return c.JSON(options)
}

func AfterRegistration(c *fiber.Ctx) error {
    name := c.Params("name", "unkown")
    user := users[name]
    if user == nil {
        return fiber.ErrUnauthorized
    }
    
    println("-", len(users), "-")
    if user.Session == nil {
        return fiber.ErrUnauthorized
    }

    r, err := adaptor.ConvertRequest(c, true)
    if err != nil {
        return fiber.ErrInternalServerError
    }

    credential, err := webAuthn.FinishRegistration(user, *user.Session, r)
    if err != nil {
        return err
    }

    user.Credentials = append(user.Credentials, *credential)
    return nil
}
