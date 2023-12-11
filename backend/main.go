package main

import (
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	fiber "github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/google/uuid"
	"google.golang.org/genproto/googleapis/iam/credentials/v1"
)

var webAuthn *webauthn.WebAuthn //not proud of how I initialize this
var appName = "Go Fido2"
var (
    users = make(map[string]*User)
)

type RegistrationResult struct {
    CredentialId string `json:"credential_id"`
    UserId string `json:"user_id"`
}

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
    var p int
    flag.IntVar(&p, "port", 8080, "Specify the port number")
    flag.Parse()
   var port string = ":" + string(port) 

    wconfig := &webauthn.Config{
            RPDisplayName: appName, // Display Name for your site
        RPID: "fido2.com", // Generally the FQDN for your site
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
	
    app.Get("/", func(c *fiber.Ctx) error {
        return c.SendString("Hello, World!")
    })

    app.Get("/register/:name", func(c *fiber.Ctx) error {
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

    app.Get("/authenticate/:name", func(c *fiber.Ctx) error {
        err := BeforeAuthenticate(c) 
        if err != nil {
            return err
        }
        err = AfterAuthenticate(c)
        if err != nil {
            return err
        }
        return nil
    })

    app.Get("/register/:name/begin", BeforeRegistration)
    app.Post("/register/:name/end", AfterRegistration)

    app.Get("/authenticate/:name/begin", BeforeAuthenticate)
    app.Get("/authenticate/:name/end", AfterAuthenticate)
    app.Get("/handlers", func(c *fiber.Ctx)error{
       return c.JSON(app.GetRoutes())
    })

    //if err := app.ListenTLS(p, "./assets/cert.pem", "./assets/key.pem");err != nil {
    if err := app.Listen(p);err != nil{
        panic(err)
    }

}

func BeforeAuthenticate(c *fiber.Ctx) error {
    name := c.Params("name","unknown")
    user := getUser(name)

    options, session, err := webAuthn.BeginAuthenticate(user)
    if err != nil {
        return fiber.ErrInternalServerError
    }

    user.Session = session
    users[user.Name] = &user // do not understand why this is neccessary

    return c.JSON(options)
}

func AfterAuthenticate(c *fiber.Ctx) error {
    name := c.Params("name", "unkown")
    user := users[name]
    if user == nil {
        return fiber.ErrUnauthorized
    }
    
    if user.Session == nil {
        return fiber.ErrUnauthorized
    }

    r, err := adaptor.ConvertRequest(c, true)
    if err != nil {
        return fiber.ErrInternalServerError
    }

    credential, err := webAuthn.FinishAuthenticate(user, *user.Session, r)
    if err != nil {
        return err
    }

    for _, cred := range user.Credentials {
        if string(cred.ID) == string(credential.ID) {
            return c.JSON(fiber.StatusOK)
        }
    }
    return fiber.ErrUnauthorized
}

func BeforeRegistration(c *fiber.Ctx) error {
    name := c.Params("name","unknown")
    user := getUser(name)

//    t := true
    options, session, err := webAuthn.BeginRegistration(
		user,
        /*
		webauthn.WithAuthenticatorSelection(protocol.AuthenticatorSelection{
			RequireResidentKey: &t,
			ResidentKey:        protocol.ResidentKeyRequirementRequired,
			UserVerification:   protocol.UserVerificationRequirement(h.cfg.Webauthn.UserVerification),
		}),
        */
		webauthn.WithConveyancePreference(protocol.PreferNoAttestation),
		// don't set the excludeCredentials list, so an already registered device can be re-registered
	)
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

    res := RegistrationResult{
        UserId: user.Id,
        CredentialId: string(credential.ID),
    }

    return c.JSON(res)
}

