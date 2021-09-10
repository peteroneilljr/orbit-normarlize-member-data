package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/tidwall/gjson"
)

var orbitApiKey = os.Getenv("ORBIT_API_KEY")
var orbitWorkspaceID = os.Getenv("ORBIT_WORKSPACE_ID")
var orbitField string
var validSearch == false = true
var newName string
var numMembersChanged = 0

func updateMember(memberID string, changeField string, newData string) {

	url := fmt.Sprintf("https://app.orbit.love/api/v1/%s/members/%s", orbitWorkspaceID, memberID)

	payloadString := fmt.Sprintf("{\"%s\":\"%s\"}", changeField, newData)
	payload := strings.NewReader(payloadString)

	req, err := http.NewRequest("PUT", url, payload)
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", orbitApiKey)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer res.Body.Close()
	if res.StatusCode == 204 {
		fmt.Printf("%s changed to: %s", orbitField, newName)
		numMembersChanged++
	} else {
		fmt.Println("Error: HTTP Status Code:", res.StatusCode)
	}
}
func getMemberList(orbitField string, oldName string) []byte {
	url := fmt.Sprintf("https://app.orbit.love/api/v1/%s/members?%s=%s&items=100", orbitWorkspaceID, orbitField, url.PathEscape(oldName))

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err)
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", orbitApiKey)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
	}

	if res.StatusCode != 200 {
		fmt.Println("Error finding members: HTTP Status Code:", res.StatusCode)
		os.Exit(1)
	}

	return body
}

func main() {

	flag.StringVar(&orbitField, "field", "", "The field in Orbit you wish to update")
	flag.StringVar(&newName, "new", "", "This will replace the old data")

	flag.Parse()

	oldNames := flag.Args()
	fmt.Println("Orbit Field:", orbitField)
	fmt.Println("New Name:", newName)
	fmt.Println("Old Names:", oldNames)

	if orbitField == "" {
		fmt.Println("Please provided the field you wish to scan with: --field")
		validSearch = false
	}
	if len(oldNames) == 0 {
		fmt.Println("Please provide the names you want to search for after command line flags")
		validSearch = false
	}
	if orbitApiKey == "" {
		fmt.Println("Please provided an API key using env var ORBIT_API_KEY")
		validSearch = false
	}
	if orbitWorkspaceID == "" {
		fmt.Println("Please provided your Orbit Workspace ID using env var ORBIT_WORKSPACE_ID")
		validSearch = false
	}
	if validSearch == false {
		os.Exit(1)
	}

	for i := 0; i < len(oldNames); i++ {

		fmt.Printf("Checking for members with %s: %s\n", orbitField, oldNames[i])

		body := getMemberList(orbitField, oldNames[i])
		memberIDs := gjson.GetBytes(body, "data.#.attributes.id")
		if memberIDs.Raw == "[]" {
			fmt.Println("No members found")
			continue
		}

		matches := len(memberIDs.Array())
		fmt.Println("Number of matching members:", matches)

		if newName != "" {
			for x := 0; x < matches; x++ {

				memberName := gjson.GetBytes(body, fmt.Sprintf("data.%d.attributes.name", x))
				fmt.Println("Member Name:", memberName)

				memberEmail := gjson.GetBytes(body, fmt.Sprintf("data.%d.attributes.email", x))
				fmt.Println("Member Email:", memberEmail)

				memberID := gjson.GetBytes(body, fmt.Sprintf("data.%d.attributes.id", x))
				fmt.Println("Member ID:", memberID)

				updateMember(memberID.String(), orbitField, newName)
			}
		}
	}

	fmt.Printf("\n===\nNumber Of Members Updated: %d\n", numMembersChanged)
}
