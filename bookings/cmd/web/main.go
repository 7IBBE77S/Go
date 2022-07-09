package main

import (
	"bufio"
	"fmt"
	"strconv"

	// "html/template"
	"log"
	"net/http"
	"os"

	// "strconv"
	"strings"
	"time"

	"github.com/7IBBE77S/web-app/pkg/config"
	"github.com/7IBBE77S/web-app/pkg/handlers"
	"github.com/7IBBE77S/web-app/pkg/render"
	"github.com/alexedwards/scs/v2"
)

type portColor struct {
	BrightRed string
	Red       string
	Green     string
	Yellow    string
	Blue      string
	Magenta   string
	Purple    string
	Cyan      string
	White     string
	Grey      string
	Clean     string
}

var colors = portColor{
	BrightRed: "\033[1;31m",
	Red:       "\x1b[31m",
	Green:     "\x1b[32m",
	Yellow:    "\x1b[33m",
	Blue:      "\x1b[34m",
	Magenta:   "\x1b[35m",
	Purple:    "\x1b[36m",
	Cyan:      "\x1b[36m",
	White:     "\x1b[37m",
	Grey:      "\x1b[90m",
	Clean:     "\x1b[0m",
}
var reader = bufio.NewReader(os.Stdin)

var port, _ = getPortInput()
var app config.AppConfig
var session *scs.SessionManager

//TODO: REMOVE 3RD PARTY PACKAGES AND REPLACE WITH OWN CODE!
func main() {
	//In production this needs to be set to true.
	app.InProduction = false

	session = scs.New()
	session.Lifetime = time.Hour * 24
	session.Cookie.Persist = true

	session.Cookie.SameSite = http.SameSiteLaxMode
	//In production this needs to be set to true.
	session.Cookie.Secure = app.InProduction

	app.Session = session

	tc, err := render.CreateTemplateCache()

	if err != nil {
		log.Fatal(err)
	}
	app.TemplateCache = tc
	app.UseCache = false
	repo := handlers.NewRepo(&app)
	handlers.NewHandler(repo)
	render.NewTemplates(&app)
	startPort(port)

}
func getPortInput() (string, error) {
	fmt.Print("Enter a port number -> ")

	userInput, _ := reader.ReadString('\n')
	userInput = strings.Replace(userInput, " ", "", -1)
	userInput = strings.Replace(userInput, "\n", "", -1)

	if userInput == "" {
		fmt.Println(colors.Red + "Error: " + colors.Clean + "Port cannot be " + colors.Blue + "nil" + colors.Clean + ".")
		time.Sleep(1 * time.Second)
		userInput = "8080"
		fmt.Println("-> Defaulting to " + colors.Blue + userInput + colors.Clean + ".")
	}
	//Check if port enter is a number

	_, err := strconv.Atoi(userInput)
	if err == nil {
		if strings.HasPrefix(userInput, "0") {
			fmt.Println(colors.BrightRed + "Error: " + colors.Clean + "Port cannot begin with 0.")

			userInput = strings.Replace(userInput, "0", "1", 1)
			fmt.Println("-> Port has been updated to " + colors.Blue + userInput + colors.Clean + ".")
		}
		time.Sleep(1 * time.Second)
		if len(userInput) > 4 {
			userInput = userInput[:4]
			fmt.Println("-> Port has been updated to " + colors.Blue + userInput + colors.Clean + ".")

		} else if len(userInput) < 4 {
			for len(userInput) <= 4 {
				userInput += "0"
			}
			fmt.Println("-> Port has been updated to " + colors.Blue + userInput + colors.Clean + ".")
		}
		time.Sleep(1 * time.Second)
		if userInput <= "1023" {
			fmt.Println(colors.Yellow + "Warning: " + colors.Clean + "Port entered " + colors.Grey + userInput + colors.Clean + "." + " Ports less than 1024 or greater than 9999 are protected.")
			time.Sleep(1 * time.Second)

			userInput = "8080"
			fmt.Println("-> Defaulting to " + colors.Blue + userInput + colors.Clean + ".")
		}
	} else if err != nil {
		for {
			return getPortInput()
		}
	}

	return userInput, nil
}

func startPort(port string) {

	fmt.Println()
	portColor := colors.Green + port + colors.Clean

	log.Printf("Server started on port %s.", portColor)

	// http.HandleFunc("/", handlers.Repo.Home)
	// http.HandleFunc("/about", handlers.Repo.About)

	// _ = http.ListenAndServe("127.0.0.1:"+port, nil)

	srv := &http.Server{
		Addr:    "127.0.0.1:" + port,
		Handler: routes(handlers.Repo.App),
	}
	err := srv.ListenAndServe()

	log.Fatal(err)

}
