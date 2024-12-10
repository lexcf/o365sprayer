<p align="center">
<img src="static/O365Sprayer.png" width="50%">
</p>

<h4 align="center">Enumerate & Spray O365 Accounts.</h4>

<p align="center">
<a href="https://github.com/securebinary/o365sprayer/"><img src="https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat"></a>
<a href="https://twitter.com/thesecurebinary"><img src="https://img.shields.io/twitter/follow/thesecurebinary.svg?logo=twitter"></a>
</p>
      
<p align="center">
  <a href="#features">Features</a> •
  <a href="#installation">Install</a> •
  <a href="#usage">Usage</a>
</p>

## Features

 - Distinguishes Managed O365 & Federated O365 for the target domain
 - Enumerates emails for valid O365 accounts
 - Sprays passwords to check for valid credentials
 - Provide custom delay between each request
 - Provide number of attempts which triggers account lockout
 - Provide cool down time for account lockout
 - Provide maximum number of account lockouts to tolerate while spraying
 - Add multithreading

## Installation

```bash
git clone https://github.com/lexcf/o365sprayer.git
cd o365sprayer
go run o365sprayer.go
```

## Usage

```bash
aidenpearce369@horus~ go run o365sprayer.go

   ____                              ___    _
  / __/ ___  ____ __ __  ____ ___   / _ )  (_)  ___  ___ _  ____  __ __
 _\ \  / -_)/ __// // / / __// -_) / _  | / /  / _ \/ _  / / __/ / // /
/___/  \__/ \__/ \_,_/ /_/   \__/ /____/ /_/  /_//_/\_,_/ /_/    \_, /
                                                                /___/
             O365 Sprayer v1.0.1 (fork by @lexcf, multithread)

  -d
      Target domain
  -u
      Email to validate
  -p
      Password to spray
  -U
      Path to email list
  -P
      Path to password list
  -enum [DEFAULT : false]
      Validate O365 emails
  -spray [DEFAULT : false]
      Spray passwords on O365 emails
  -delay [DEFAULT : 0.25]
      Delay between requests
  -lockout [DEFAULT : 5]
      Number of incorrect attempts for account lockout
  -lockoutDelay [DEFAULT : 15]
      Lockout cool down time
  -max-lockout [DEFAULT : 10]
      Maximum number of lockout accounts
  -t [DEFAULT : 40]
      Number of threads to use

```

This will display help for the CLI tool. Here are all the required arguments it supports.

## License
`O365 Sprayer` is made with 🖤 by the `SecureBinary` team. Any tweaks / community contribution are welcome.
