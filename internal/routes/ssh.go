package routes

import (
	"errors"
	"net"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/elapsed"
	"github.com/charmbracelet/wish/logging"
	"github.com/jfosburgh/gomes/internal/routes/models"
)

const (
	host = "localhost"
	port = "23234"
)

func ServeSSH() {
	srv, err := wish.NewServer(
		// The address the server will listen to.
		wish.WithAddress(net.JoinHostPort(host, port)),

		// The SSH server need its own keys, this will create a keypair in the
		// given path if it doesn't exist yet.
		// By default, it will create an ED25519 key.
		wish.WithHostKeyPath(".ssh/id_ed25519"),
		// Middlewares do something on a ssh.Session, and then call the next
		// middleware in the stack.
		wish.WithMiddleware(
			bubbletea.Middleware(teaHandler),
			activeterm.Middleware(), // Bubble Tea apps usually require a PTY.

			// The last item in the chain is the first to be called.
			logging.Middleware(),
			elapsed.Middleware(),
		),
	)
	if err != nil {
		log.Error("Could not start server", "error", err)
	}

	log.Info("Starting SSH server", "host", host, "port", port)
	go func() {
		if err = srv.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			// We ignore ErrServerClosed because it is expected.
			log.Error("Could not start server", "error", err)
		}
	}()
}

// You can wire any Bubble Tea model up to the middleware with a function that
// handles the incoming ssh.Session. Here we just grab the terminal info and
// pass it to the new model. You can also return tea.ProgramOptions (such as
// tea.WithAltScreen) on a session by session basis.
func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	// This should never fail, as we are using the activeterm middleware.
	pty, _, _ := s.Pty()

	// When running a Bubble Tea app over SSH, you shouldn't use the default
	// lipgloss.NewStyle function.
	// That function will use the color profile from the os.Stdin, which is the
	// server, not the client.
	// We provide a MakeRenderer function in the bubbletea middleware package,
	// so you can easily get the correct renderer for the current session, and
	// use it to create the styles.
	// The recommended way to use these styles is to then pass them down to
	// your Bubble Tea model.
	renderer := bubbletea.MakeRenderer(s)
	txtStyle := renderer.NewStyle().Foreground(lipgloss.Color("10"))
	quitStyle := renderer.NewStyle().Foreground(lipgloss.Color("8"))

	m := models.ModelHome{
		WindowParams: models.WindowParams{
			Width:     pty.Window.Width,
			Height:    pty.Window.Height,
			TxtStyle:  txtStyle,
			QuitStyle: quitStyle,
		},
		Games: []string{
			"Chess",
			"Tic-Tac-Toe",
		},
	}
	return m, []tea.ProgramOption{tea.WithAltScreen()}
}
