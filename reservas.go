package main


import (
        "github.com/prometheus/client_golang/prometheus"
        "github.com/prometheus/client_golang/prometheus/promhttp"
        "html/template"
        "log"
        "net/http"
        "os"
        "encoding/csv"
        "strconv"
)

type Sala struct {
        Name     string
        Reservas []Reserva
}

type Reserva struct {
        Sala  string
        Dia   string
        Hora  string
        User  string
}

type User struct {
        Username string
        Password string
        Admin    bool
}

type PageData struct {
        Adm bool
}

var (
    reservationsCounter = prometheus.NewCounter(prometheus.CounterOpts{
        Name: "numero_reservas",
        Help: "Numero de reservas feitas",
    })
)

var (
        salas  []Sala
        users  []User
        admin  User
        adm    bool
)

func main() {
        salas = make([]Sala, 0)
        users = make([]User, 0)

        le_Salas("dados/salas.csv")
        le_Users("dados/users.csv")
        le_Reservas ()

        http.HandleFunc("/", handleLogin)
        http.HandleFunc("/login", handleLogin)
        http.HandleFunc("/logout", handleLogout)
        http.HandleFunc("/reserva", handleReserva)
        http.HandleFunc("/cancela", handleCancel)
        http.HandleFunc("/status", handleStatus)
        http.HandleFunc("/menu", handleMenu)

        prometheus.MustRegister(reservationsCounter)
        http.Handle("/metrics", promhttp.Handler())

        log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleMenu(w http.ResponseWriter, r *http.Request) {
        if !Logged(w, r) {
                http.Redirect(w, r, "/login", http.StatusSeeOther)
                return
        }

        username, _ := r.Cookie("username")

        data := PageData{
            Adm: isAdmin(username.Value),
        }

        tmpl := template.Must(template.ParseFiles("menu.html"))
        tmpl.Execute(w, data)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
        if Logged(w, r) {
                http.Redirect(w, r, "/menu", http.StatusSeeOther)
                return
        }

        tmpl := template.Must(template.ParseFiles("login.html"))

        if r.Method == http.MethodPost {
                username := r.FormValue("username")
                password := r.FormValue("password")

                for _, user := range users {
                        if user.Username == username && user.Password == password {
                                http.SetCookie(w, &http.Cookie{
                                        Name:  "username",
                                        Value: username,
                                })
                                http.Redirect(w, r, "/menu", http.StatusSeeOther)
                                return
                        }
                }

                http.Redirect(w, r, "/", http.StatusSeeOther)
                return
        }

        tmpl.Execute(w, nil)
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
        if !Logged(w, r) {
                http.Redirect(w, r, "/login", http.StatusSeeOther)
                return
        }

        http.SetCookie(w, &http.Cookie{
                Name:   "username",
                Value:  "",
                MaxAge: -1,
        })
        http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleReserva(w http.ResponseWriter, r *http.Request) {
        if !Logged(w, r) {
                http.Redirect(w, r, "/login", http.StatusSeeOther)
                return
        }

        if r.Method == http.MethodPost {
                username, err := r.Cookie("username")
                if err != nil {
                        http.Redirect(w, r, "/", http.StatusSeeOther)
                        return
                }

                if username.Value == "" {
                        http.Redirect(w, r, "/", http.StatusSeeOther)
                        return
                }

                sala := r.FormValue("sala")
                dia := r.FormValue("dia")
                hora := r.FormValue("hora")

                for i, room := range salas {
                        if room.Name == sala {
                                salas[i].Reservas = append(room.Reservas, Reserva{
                                        Sala:  sala,
                                        Dia:   dia,
                                        Hora:  hora,
                                        User:  username.Value,
                                })
                                break
                        }
                }

                reservationsCounter.Inc()
                salva_Reservas ()
                http.Redirect(w, r, "/status", http.StatusSeeOther)
                return
        }

        tmpl := template.Must(template.ParseFiles("reserva.html"))
        tmpl.Execute(w, salas)
}

func handleCancel(w http.ResponseWriter, r *http.Request) {
        if !Logged(w, r) {
                http.Redirect(w, r, "/login", http.StatusSeeOther)
                return
        }

        username, _ := r.Cookie("username")
        if !isAdmin(username.Value) {
                http.Redirect(w, r, "/status", http.StatusSeeOther)
                return
        }

        if r.Method == http.MethodPost {
                sala := r.FormValue("sala")
                dia := r.FormValue("dia")
                hora := r.FormValue("hora")

                for i, room := range salas {
                        if room.Name == sala {
                                for j, reserve := range room.Reservas {
                                        if reserve.Sala == sala && reserve.Dia == dia && reserve.Hora == hora {
                                                salas[i].Reservas = append(room.Reservas[:j], room.Reservas[j+1:]...)
                                                break
                                        }
                                }
                                break
                        }
                }

                salva_Reservas ()
                http.Redirect(w, r, "/status", http.StatusSeeOther)
                return
        }

        tmpl := template.Must(template.ParseFiles("cancela.html"))
        tmpl.Execute(w, salas)
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
        if !Logged(w, r) {
                http.Redirect(w, r, "/login", http.StatusSeeOther)
                return
        }

        tmpl := template.Must(template.ParseFiles("status.html"))
        tmpl.Execute(w, salas)
}

func isAdmin(username string) bool {
        for _, user := range users {
                if user.Username == username && user.Admin {
                        return true
                }
        }
        return false
}

func le_Salas(arquivo string) (error) {
        file, err := os.Open(arquivo)
        if err != nil {
                return err
        }
        defer file.Close()

        reader := csv.NewReader(file)
        lines, err := reader.ReadAll()
        if err != nil {
                return err
        }

        for _, line := range lines {
                sala := Sala {
                   Name: line[0],
                   Reservas: []Reserva{},
                }
                salas = append(salas, sala)
        }

        return nil
}

func le_Users (arquivo string) (error) {
        file, err := os.Open(arquivo)
        if err != nil {
                return err
        }
        defer file.Close()

        reader := csv.NewReader(file)
        lines, err := reader.ReadAll()
        if err != nil {
                return err
        }

        for _, line := range lines {
                b, _ := strconv.ParseBool(line[2])
                usuario := User {
                   Username: line[0],
                   Password: line[1],
                   Admin: b,
                }
                users = append(users, usuario)
        }

        return nil
}

func le_Reservas () (error) {
        file, err := os.Open("dados/reservas.csv")
        if err != nil {
                return err
        }
        defer file.Close()

        reader := csv.NewReader(file)
        lines, err := reader.ReadAll()
        if err != nil {
                return err
        }


        for _, line := range lines {
                for i, room := range salas {
                        if room.Name == line[0] {
                                salas[i].Reservas = append(room.Reservas, Reserva{
                                        Sala:  line[0],
                                        Dia:   line[1],
                                        Hora:  line[2],
                                        User:  line[3],
                                })
                                reservationsCounter.Inc()
                                break
                        }
                }
        }

        return nil
}

func salva_Reservas () error {
        file, err := os.Create("dados/reservas.csv")
        if err != nil {
                return err
        }
        defer file.Close()

        writer := csv.NewWriter(file)
        defer writer.Flush()


        // Escrever dados das reservas
        for _, room := range salas {
                for _, reserva := range room.Reservas {
                        row := []string{room.Name, reserva.Dia, reserva.Hora, reserva.User}
                        err = writer.Write(row)
                        if err != nil {
                                return err
                        }
                }
        }

        return nil
}

func Logged (w http.ResponseWriter, r *http.Request) bool {
        username, err := r.Cookie("username")
        if err != nil {
                return false
        }

        if username.Value == "" {
                return false
        }

        return true
}

