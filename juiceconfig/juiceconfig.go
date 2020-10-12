/*juiceconfig Abstract configurations between JSON file, Secrets Manager, and environment variables.
 */
package juiceconfig

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/antonholmquist/jason"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

const (
	FILE_PREFIX            = "file:::"
	SECRETS_MANAGER_PREFIX = "secrets_manager:::"
	ENVIRONMENT_PREFIX     = "environment:::"
)

/*JuiceConfig Configuration object loading using a URL.
 */
type JuiceConfig struct {
	config   *map[string]*jason.Value
	URL      string
	hadError bool
	errMsg   string
}

/*Load Load configuration
 */
func Load(url string) (*JuiceConfig, error) {
	// fmt.Printf("juiceconfig.Load(%s)\n", url)
	obj := &JuiceConfig{}
	obj.URL = url

	var data []byte
	var err error
	if strings.HasPrefix(url, FILE_PREFIX) {

		// Load configuration from a file
		filename := url[len(FILE_PREFIX):]
		data, err = ioutil.ReadFile(filename)
		if err != nil {
			fmt.Printf("Unable to open config file %s\n", filename)
			return obj, obj.setError("Unable to open config file [" + filename + "]")
		}

	} else if strings.HasPrefix(url, SECRETS_MANAGER_PREFIX) {

		// Get the configuration from AWS Secrets Manager
		def := url[len(SECRETS_MANAGER_PREFIX):]
		// Split into region:::secretName
		pos := strings.Index(def, ":::")
		if pos < 0 {
			fmt.Printf("Invalid URL for JuiceConfig: %s\n", url)
			return obj, obj.setError("Invalid URL [" + url + "]")
		}
		region := def[0:pos]
		secretName := def[pos+3:]
		// Connect to AWS
		sess := session.Must(session.NewSession(&aws.Config{
			Region: &region,
		}))
		svc := secretsmanager.New(sess)
		params := &secretsmanager.GetSecretValueInput{
			SecretId:     aws.String(secretName),
			VersionStage: aws.String("AWSCURRENT"),
		}
		// Get the secret
		result, err := svc.GetSecretValue(params)
		if err != nil {
			fmt.Printf("Unable to access AWS Secrets Manager: %+v\n", err)
			return obj, obj.setError("Unable to access AWS Secrets Manager [" + err.Error() + "]")
		}
		secret := result.SecretString
		data = []byte(*secret)

	} else if strings.HasPrefix(url, ENVIRONMENT_PREFIX) {

		// Get the config from an environment variable
		variableName := url[len(ENVIRONMENT_PREFIX):]
		envvar := os.Getenv(variableName)
		if envvar == "" {
			fmt.Println("Environment variable not set [" + variableName + "]")
			return obj, obj.setError("Environment variable not set [" + variableName + "]")
		}
		data = []byte(envvar)

	} else {

		// Unknown URL prefix.
		fmt.Printf("Invalid URL for JuiceConfig: %s\n", url)
		return obj, obj.setError("Invalid URL for JuiceConfig [" + url + "]")
	}

	// Parse the configuration
	config, err := jason.NewObjectFromBytes(data)
	if err != nil {
		fmt.Printf("Error parsing config file: %+v\n", err)
		return obj, obj.setError("Error parsing config file [" + err.Error() + "]")
	}

	// Flatten the config
	newConfig := map[string]*jason.Value{}
	err = flattenConfig(&newConfig, "", config)
	if err != nil {
		fmt.Printf("Error flattening config: ", err)
		return obj, obj.setError("Error flattening config [" + err.Error() + "]")
	}
	obj.config = &newConfig

	// All good
	obj.hadError = false
	return obj, nil
}

/*flattenConfig Recursively flatten a configuration.
 *
 *	Configs may specify nested values, for example:
 *		"app": {
 *				"name": "myApp",
 *				"size": 10
 *		}
 *
 *	This function receusively flattens it like this into a map:
 *		"app.name": "myApp",
 *		"app.size": 10
 */
func flattenConfig(newConfig *(map[string](*jason.Value)), prefix string, config *jason.Object) error {

	for key, value := range config.Map() {
		path := prefix + key

		// See if it is an Object
		obj, objErr := value.Object()
		if objErr == jason.ErrNotObject {

			// Save this value
			(*newConfig)[path] = value
		} else if objErr == nil {

			// Add the values within this object
			flattenConfig(newConfig, path+".", obj)
		}
	}
	return nil
}

/*WasError Has an error occurred?
 *	Rather than checking for errors every time we get a config value,
 *	we can check at the end.
 */
func (jc *JuiceConfig) WasError() bool {
	return jc.hadError
}

/*ErrorMessage Return description of previous error
 */
func (jc *JuiceConfig) ErrorMessage() string {
	return jc.errMsg
}

/*ResetError Reset the error stat.s
 */
func (jc *JuiceConfig) ResetError() {
	jc.hadError = false
}

/*setError
 *	Set error status, remember the message, and return an Error object
 */
func (jc *JuiceConfig) setError(msg string) error {
	jc.hadError = true
	jc.errMsg = msg
	return errors.New(msg)
}

/*GetString Get a string configuration value
 */
func (jc *JuiceConfig) GetString(path string, dflt ...string) (string, error) {
	if jc.hadError {
		return "", errors.New("Already had error")
	}
	jvalue, ok := (*jc.config)[path]
	if ok {
		value, err := jvalue.String()
		if err == nil {
			return value, nil
		}
		return "", jc.setError("Value is not string [" + path + "]")
	}

	// Value not found. Is there a default?
	if len(dflt) > 0 {
		return dflt[0], nil
	}
	return "", jc.setError("Value not found [" + path + "]")
}

/*GetInt Get an integer configuration value
 */
func (jc *JuiceConfig) GetInt(path string, dflt ...int64) (int64, error) {
	if jc.hadError {
		return 0, errors.New("Already had error")
	}
	jvalue, ok := (*jc.config)[path]
	if ok {
		value, err := jvalue.Int64()
		if err == nil {
			return value, nil
		}
		return 0, jc.setError("Value is not int64 [" + path + "]")
	}

	// Value not found. Is there a default?
	if len(dflt) > 0 {
		return dflt[0], nil
	}
	return 0, jc.setError("Value not found [" + path + "]")
}

/*GetBool Get an integer configuration value
 */
func (jc *JuiceConfig) GetBool(path string, dflt ...bool) (bool, error) {
	if jc.hadError {
		return false, errors.New("Already had error")
	}
	jvalue, ok := (*jc.config)[path]
	if ok {
		value, err := jvalue.Boolean()
		if err == nil {
			return value, nil
		}
		return false, jc.setError("Value is not bool [" + path + "]")
	}

	// Value not found. Is there a default?
	if len(dflt) > 0 {
		return dflt[0], nil
	}
	return false, jc.setError("Value not found [" + path + "]")
}

/*
 *	Conveniance functions for simpler access.
 *	A default configuration file is defined using environment variable JUICE_CONFIG.
 */
var defaultConfig *JuiceConfig

func checkDefaultConfigIsLoaded() error {

	// Is the default config already loaded?
	if defaultConfig != nil {
		return nil
	}

	// Get the config location from an environment variable
	envvar := os.Getenv("JUICE_CONFIG")
	fmt.Printf("JUICE_CONFIG=%s.\n", envvar)

	// Load the default config
	var err error
	defaultConfig, err = Load(envvar)
	if err != nil {
		fmt.Printf("Error loading default config: %+v\n", err)
		return err
	}
	return nil
}

/*GetString Get a string value from the config defined by JUICE_CONFIG.
 */
func GetString(path string, dflt ...string) (string, error) {
	checkDefaultConfigIsLoaded()
	return defaultConfig.GetString(path, dflt...)
}

/*GetInt Get an integer value from the config defined by JUICE_CONFIG.
 */
func GetInt(path string, dflt ...int64) (int64, error) {
	checkDefaultConfigIsLoaded()
	return defaultConfig.GetInt(path, dflt...)
}

/*GetBool Get a boolean value from the config defined by JUICE_CONFIG.
 */
func GetBool(path string, dflt ...bool) (bool, error) {
	checkDefaultConfigIsLoaded()
	return defaultConfig.GetBool(path, dflt...)
}

/*WasError Has an error occurred?
 *	Rather than checking for errors every time we get a config value,
 *	we can check at the end.
 */
func WasError() bool {
	return defaultConfig.WasError()
}

/*ErrorMessage Get the previous error's message
 */
func ErrorMessage() string {
	return defaultConfig.ErrorMessage()
}

/*ResetError Reset the error status.
 */
func ResetError() {
	defaultConfig.ResetError()
}
