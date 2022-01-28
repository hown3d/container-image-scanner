package testutil

import "fmt"

func GenerateTestRegistryJSON(registryScheme bool, domain, username, password string) string {
	if registryScheme {
		return fmt.Sprintf(`
		{
			"auths":{
				"%v":{
					"username":"%v",
					"password":"%v"
				}
			}
		}`, domain, username, password)
	}
	return fmt.Sprintf(`
	{
		"username":"%v",
		"password":"%v"
	}
	`, username, password)
}
