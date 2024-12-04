package main

import (
    "bufio"
    "fmt"
    "io"
    "log"
    "net"
    "net/http"
    "os"
    "os/exec"
    "strings"
	"encoding/json"
)

type PIDs struct {
    LaravelPID int `json:"laravel_pid"`
    NpmPID     int `json:"npm_pid"`
}

func main() {
    reader := bufio.NewReader(os.Stdin)

    // Obtém o IP local da máquina
    ip, err := getLocalIP()
    if err != nil {
        log.Fatalf("Erro ao obter o IP da máquina: %v", err)
    }
    fmt.Printf("IP detectado: %s\n", ip)

    // Perguntar ao usuário a porta
    fmt.Print("Digite a porta para servir a aplicação (ex: 8080): ")
    port, _ := reader.ReadString('\n')
    port = strings.TrimSpace(port)

    address := fmt.Sprintf("%s:%s", ip, port)

    fmt.Printf("A aplicação será servida em %s:%s\n", ip, port)

    // Iniciar o servidor Laravel com php artisan serve no IP detectado
    laravelCmd := exec.Command("php", "artisan", "serve", fmt.Sprintf("--host=%s", ip), fmt.Sprintf("--port=%s", port))
    laravelCmd.Dir = "./src"
    laravelCmd.Stdout = os.Stdout
    laravelCmd.Stderr = os.Stderr
    if err := laravelCmd.Start(); err != nil {
        log.Fatalf("Erro ao iniciar o servidor Laravel: %v", err)
    }

    // Iniciar o servidor npm com npm run dev
    npmCmd := exec.Command("npm", "run", "dev")
    npmCmd.Dir = "./src"
    npmCmd.Stdout = os.Stdout
    npmCmd.Stderr = os.Stderr
    if err := npmCmd.Start(); err != nil {
        log.Fatalf("Erro ao iniciar o npm: %v", err)
    }

    // Salvar os PIDs dos processos em um arquivo JSON
    pids := PIDs{
        LaravelPID: laravelCmd.Process.Pid,
        NpmPID:     npmCmd.Process.Pid,
    }

    savePIDsToFile(pids, "pids.json")

    // Usar Go como proxy reverso para o Laravel
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        laravelURL := fmt.Sprintf("http://%s:%s%s", ip, port, r.URL.Path)
        resp, err := http.Get(laravelURL)
        if err != nil {
            http.Error(w, "Erro ao acessar o Laravel", http.StatusBadGateway)
            return
        }
        defer resp.Body.Close()

        // Copiar o conteúdo da resposta do Laravel para o cliente
        for key, value := range resp.Header {
            w.Header().Set(key, strings.Join(value, ","))
        }
        w.WriteHeader(resp.StatusCode)
        io.Copy(w, resp.Body)
    })

    log.Printf("Servidor rodando em http://%s\n", address)
    log.Fatal(http.ListenAndServe(address, nil))
}

// Função para obter o IP local da máquina
func getLocalIP() (string, error) {
    addrs, err := net.InterfaceAddrs()
    if err != nil {
        return "", err
    }

    for _, addr := range addrs {
        var ip net.IP
        switch v := addr.(type) {
        case *net.IPNet:
            ip = v.IP
        case *net.IPAddr:
            ip = v.IP
        }
        if ip != nil && !ip.IsLoopback() && ip.To4() != nil {
            return ip.String(), nil
        }
    }
    return "", fmt.Errorf("não foi possível encontrar um IP local")
}

func savePIDsToFile(pids PIDs, filename string) error {
    file, err := os.Create(filename)
    if err != nil {
        return fmt.Errorf("erro ao criar o arquivo de PIDs: %v", err)
    }
    defer file.Close()

    encoder := json.NewEncoder(file)
    if err := encoder.Encode(pids); err != nil {
        return fmt.Errorf("erro ao salvar os PIDs no arquivo: %v", err)
    }

    log.Printf("PIDs salvos no arquivo %s\n", filename)
    return nil
}
