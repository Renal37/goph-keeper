package core

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Renal37/goph-keeper/internal/agent/client"
)

var defaultPermition fs.FileMode = 0600
var errorFailedReadSTDIN = "failed read stdin: %w"

func Run(client *client.Client, command string) error {
	// Depending on the command, we choose the logic of behavior
	switch command {
	case "sign-up":
		fmt.Println("-> Create new account")

		// Get user credentials from stdin tui
		ss, err := getUserCredentials()
		if err != nil {
			return fmt.Errorf("failed get user credentials: %w", err)
		}

		r, err := client.Register(ss.login, ss.password)
		if err != nil {
			return fmt.Errorf("failed register user: %w", err)
		}

		fmt.Printf("Token: %s \n", r.Jwt)

		// Do you want to save the token?
		err = saveAuthToken(r.Jwt)
		if err != nil {
			return fmt.Errorf("client failed save token: %w", err)
		}
	case "sign-in":
		fmt.Println("-> Sign in with your account")

		ss, err := getUserCredentials()
		if err != nil {
			return fmt.Errorf("failed get user credentials: %w", err)
		}

		r, err := client.Login(ss.login, ss.password)
		if err != nil {
			return fmt.Errorf("failed login user: %w", err)
		}

		fmt.Printf("Token: %s \n", r.Jwt)
		err = saveAuthToken(r.Jwt)
		if err != nil {
			return fmt.Errorf("client failed save token: %w", err)
		}
	case "read-file":
		fmt.Println("-> Read file")

		// Request to read all file
		rAllFile, err := client.ReadAllFile()
		if err != nil {
			return fmt.Errorf("failed get all file: %w", err)
		}

		// If there are no files, exit
		if len(rAllFile.Units) == 0 {
			fmt.Println("Not found files. Bye!")
			return nil
		}

		// Showing the available files
		fmt.Println("Available files:")
		for _, v := range rAllFile.Units {
			// TODO: Откуда 0 ? Size slice ?
			if v.Id > 0 {
				fmt.Printf("[%v] - %s \n", v.Id, v.Name)
			}
		}

		// Selecting a file to download
		i, err := selectReadFile()
		if err != nil {
			return fmt.Errorf("wrong id file: %w", err)
		}

		// Request to read the file
		rFile, err := client.ReadFile(int32(i))
		if err != nil {
			return fmt.Errorf("failed get all file: %w", err)
		}

		// If the file type is file
		if rFile.Type == "file" {
			err = saveFileInDisk(rFile.Name, rFile.Data)
			if err != nil {
				return fmt.Errorf("save file has error: %w", err)
			}
		} else {
			// Else type is text
			fmt.Println(string(rFile.Data))
		}
	case "write-file":
		fmt.Println("-> Write file")

		// Selecting the file type and the file we want to save
		err := selectWriteData(client)
		if err != nil {
			return fmt.Errorf("select write data has error: %w", err)
		}
	case "delete-file":
		fmt.Println("-> Delete file")

		// Request to read all file
		rAllFile, err := client.ReadAllFile()
		if err != nil {
			return fmt.Errorf("failed get all file: %w", err)
		}

		// If there are no files, exit
		if len(rAllFile.Units) == 0 {
			fmt.Println("Not found files. Bye!")
			return nil
		}

		// Showing the available files
		fmt.Println("Available files:")
		for _, v := range rAllFile.Units {
			// TODO: Откуда 0 ? Size slice ?
			if v.Id > 0 {
				fmt.Printf("[%v] - %s \n", v.Id, v.Name)
			}
		}

		// Select a file to delete
		i, err := selectReadFile()
		if err != nil {
			return fmt.Errorf("wrong id file: %w", err)
		}

		// Request for delete
		_, err = client.DeleteFile(int32(i))
		if err != nil {
			return fmt.Errorf("failed delete file: %w", err)
		}

		fmt.Println("File delete!")
	default:
		fmt.Printf("Command:%s not found! \n", command)
	}

	fmt.Println("Bye!")
	return nil
}

// UTILS FOR WRITE FILE.

// saveFileInDisk saving files to disk.
func saveFileInDisk(fileName string, data []byte) error {
	fmt.Println("Where do you want to save the file?")
	fmt.Print("Enter dir path: ")

	// Create a reader for input from standard input (console)
	reader := bufio.NewReader(os.Stdin)

	// Consider the user's response
	r, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf(errorFailedReadSTDIN, err)
	}

	// Trim the spaces and newline characters from the response
	dirPath := strings.TrimSpace(r)
	fullPath := filepath.Join(dirPath, fileName)

	err = os.WriteFile(fullPath, data, defaultPermition)
	if err != nil {
		return fmt.Errorf("failed write data: %w", err)
	}

	fmt.Printf("File save in: %s \n", fullPath)

	return nil
}

// selectWriteData selecting a file to download.
func selectWriteData(client *client.Client) error {
	fmt.Println("What you want send on server?")
	fmt.Println("[1] - Text")
	fmt.Println("[2] - File")
	fmt.Print("Enter a number: ")

	// Create a reader for input from standard input (console)
	reader := bufio.NewReader(os.Stdin)

	// Consider the user's response
	r, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf(errorFailedReadSTDIN, err)
	}

	// Trim the spaces and newline characters from the response
	r = strings.TrimSpace(r)

	i, err := strconv.Atoi(r)
	if err != nil {
		return fmt.Errorf("failed parse int: %w", err)
	}

	switch i {
	case 1:
		fmt.Println("What do you want to save?")
		fmt.Println("[1] - Custom text")
		fmt.Println("[2] - Login | Password")
		fmt.Println("[3] - Credit card")
		fmt.Print("Enter a number: ")

		r, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf(errorFailedReadSTDIN, err)
		}

		r = strings.TrimSpace(r)

		i, err := strconv.Atoi(r)
		if err != nil {
			return fmt.Errorf("failed parse int: %w", err)
		}

		fmt.Print("Enter name: ")

		fileName, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf(errorFailedReadSTDIN, err)
		}

		fileName = strings.TrimSpace(fileName)

		switch i {
		case 1:
			fmt.Println("Enter text:")
		//nolint:gomnd // This legal number
		case 2:
			fmt.Println("Enter loggin and password:")
		//nolint:gomnd // This legal number
		case 3:
			fmt.Println("Enter number, name, date and CVV:")
		}

		data, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf(errorFailedReadSTDIN, err)
		}

		data = strings.TrimSpace(data)

		// Send the gRPC data
		_, err = client.WriteFile("text", fileName, data)
		if err != nil {
			return fmt.Errorf("write file has error: %w", err)
		}

	//nolint:gomnd // This legal number
	case 2:
		fmt.Print("Enter the link to the file: ")

		// Consider the user's response
		filePath, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf(errorFailedReadSTDIN, err)
		}

		filePath = strings.TrimSpace(filePath)

		// Get file name
		baseName := filepath.Base(filePath)

		// Send the gRPC data
		_, err = client.WriteFile("file", baseName, filePath)
		if err != nil {
			return fmt.Errorf("write file has error: %w", err)
		}
	}

	fmt.Println("File write!")

	return nil
}

// UTILS FOR READ FILE.

// selectReadFile select a file to read.
func selectReadFile() (int, error) {
	fmt.Print("Select ID file: ")

	// Create a reader for input from standard input (console)
	reader := bufio.NewReader(os.Stdin)

	// Consider the user's response
	response, err := reader.ReadString('\n')
	if err != nil {
		return 0, fmt.Errorf("failed read stdin: %w", err)
	}

	// Trim the spaces and newline characters from the response
	response = strings.TrimSpace(response)

	// Converting the answer to a digit
	i, err := strconv.Atoi(response)
	if err != nil {
		return 0, fmt.Errorf("failed parse int: %w", err)
	}

	return i, nil
}

// UTILS FOR REGISTER AND LOGIN.

// saveAuthToken saving the token to the .env file.
func saveAuthToken(token string) error {
	fmt.Print("Do you want save token in .env? [y/N]: ")

	// Create a reader for input from standard input (console)
	reader := bufio.NewReader(os.Stdin)

	// Consider the user's response
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf(errorFailedReadSTDIN, err)
	}

	// Trim the spaces and newline characters from the response
	response = strings.TrimSpace(response)

	// Check the user's response
	if strings.ToLower(response) == "y" {
		// Open the file .env in append or create mode, if it doesn't exist yet
		file, err := os.OpenFile(".env", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, defaultPermition)
		if err != nil {
			return fmt.Errorf("failed to open .env file: %w", err)
		}

		// Write the string with the token in the format "JWT=your_token" to the file
		_, err = file.WriteString(fmt.Sprintf("JWT=%s\n", token))
		if err != nil {
			return fmt.Errorf("failed to write token to .env file: %w", err)
		}

		err = file.Close()
		if err != nil {
			return fmt.Errorf("failed close file: %w", err)
		}

		fmt.Println("Token saved in .env file.")
	}

	return nil
}

type userCredentials struct {
	login    string
	password string
}

// getUserCredentials get a pair of username and password from the user.
func getUserCredentials() (userCredentials, error) {
	fmt.Print("Enter your login: ")

	reader := bufio.NewReader(os.Stdin)

	loginResp, err := reader.ReadString('\n')
	if err != nil {
		return userCredentials{}, fmt.Errorf("failed read login stdin: %w", err)
	}

	fmt.Print("Enter your password: ")
	passwordResp, err := reader.ReadString('\n')
	if err != nil {
		return userCredentials{}, fmt.Errorf("failed read password stdin: %w", err)
	}

	return userCredentials{
		login:    strings.TrimSpace(loginResp),
		password: strings.TrimSpace(passwordResp),
	}, nil
}
