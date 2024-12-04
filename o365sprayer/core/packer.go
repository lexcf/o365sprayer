package core

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/lexcf/o365sprayer/o365sprayer/enum"
	"github.com/lexcf/o365sprayer/o365sprayer/spray"
)

func StartO365Sprayer(
	domainName string,
	validateEmail bool,
	sprayCheck bool,
	email string,
	emailFile string,
	password string,
	passwordFile string,
	delay float64,
	lockout int,
	lockoutDelay int,
	maxLockouts int,
	threads int,
) {
	enumResult := CheckO365(domainName)
	adfsCheck := false
	fmt.Println("[*] Domain Name     : " + enumResult.DomainName)
	fmt.Println("[*] Federation Name : " + enumResult.FederationBrandName)
	fmt.Println("[*] Tenant ID       : " + enumResult.TenandId)
	fmt.Println("[*] Threads		 : " + threads )
	if enumResult.NameSpaceType == "Managed" {
		color.Yellow("[+] Using Managed O365")
	}
	if enumResult.NameSpaceType == "Federated" {
		color.Yellow("[+] Using Federated O365")
		adfsCheck = true
	}
	if enumResult.NameSpaceType == "Unknown" {
		color.Yellow("[+] O365 Not Found.. Exiting!")
		os.Exit(-1)
	}
	if len(enumResult.AuthURL) > 0 {
		color.Green("[+] Found Authorization URL For Domain - " + enumResult.DomainName)
		fmt.Println("[*] Auth URL        : " + enumResult.AuthURL)
	}
	if validateEmail {
		if !adfsCheck {
			if len(email) > 0 {
				enum.EnumEmailsManagedO365(domainName, "standalone", email, "", delay, threads)
			}
			if len(emailFile) > 0 {
				enum.EnumEmailsManagedO365(domainName, "file", "", emailFile, delay, threads)
			}
		} else {
			if len(email) > 0 {
				enum.EnumEmailsADFSO365(domainName, "standalone", email, "", delay, threads)
			}
			if len(emailFile) > 0 {
				enum.EnumEmailsADFSO365(domainName, "file", "", emailFile, delay, threads)
			}
		}
	}
	if sprayCheck {
		if adfsCheck {
			adfsURL := enumResult.AuthURL
			spray.SprayEmailsADFSO365(
				domainName,
				adfsURL,
				email,
				emailFile,
				password,
				passwordFile,
				delay,
				lockout,
				lockoutDelay,
				maxLockouts,
				threads
			)
		} else {
			spray.SprayEmailsManagedO365(
				domainName,
				email,
				emailFile,
				password,
				passwordFile,
				delay,
				lockout,
				lockoutDelay,
				maxLockouts,
				threads
			)
		}
	}
}
