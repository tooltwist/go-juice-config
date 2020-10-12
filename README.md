# go-juice-config

These libraries allow a program to seamlessly switch between configuration sources (e.g. Flat Files, and Amazon Secrets Manager).

Managing configurations as an application passes from development through to production can pose challenges.

- In a development environment, you want fast, easy-to-change access to your configuration information (e.g. In a JSON file).

- In production environments you need access to configuration stored securely (e.g. AWS Secrets Manager).

- If you are using Docker, you'll want to externalise the config from your application's Docker image, so the same image can be used in multiple environments (CI, test, UAT, Staging, Production, etc).

- Environment variables are too easy for hackers to access.

To solve these problems, JuiceConfig allows an application to seamlessly switch where config information is stored. As the application is deploymented in different environments, an variable can specify where to find the configuration. The application does not need to change for the different environments, or contain configuration information, or provide switching between configurations.

This is particularly useful when deploying Docker images, because a single image can be deployed across multiple deployments, without modification. This reduces the chance of bugs being introduced when an application is deployed to production.


## JUICE_CONFIG

#### JSON File
To access the config from a JSON file, set JUICE_CONFIG to:

> file:::&lt;/path/to/config/file.json&gt;


#### AWS Secrets Manager
This is the best place to store your configuration for Amazon ECS and EC2 deployments.

> secrets_manager:::&lt;region&gt;:::&lt;secret_name&gt;

#### Environment Variable
While not ideal, this can be useful where the other forms are not usable. 

> environment:::&lt;environment_variable_name&gt;

The config is provided as JSON in an environment variable (be careful!)
  

## JSON Format
Configuration information should be stored as JSON, and JuiceConfig understands String, Integer and Boolean values.

Nested data structures are supported, and JuiceConfig flattens out the structure automatically. For example, the path `employee.address.street` could be used to access any of the following configurations.

Fully nested:  

```json
{
  "employee": {
    "address": {
      "street": "Victoria Avenue"
    }
  }
}
```

Partly nested:  
```json
{
  "employee.address": {
    "street": "Victoria Avenue"
  }
}
```
```json
{
  "employee": {
    "address.street": "Victoria Avenue"
  }
}
```

Flattened:  
```json
{
  "employee.address.street": "Victoria Avenue"
}
```



## Usage

```golang
import "github.com/tooltwist/go-juice-config/juiceconfig"
```

Configuration values can be accessed from the default configuration specified by the JUICE_CONFIG environment varaible.

```golang
stringValue, err := juiceconfig.GetString("database.hostname")
intValue, err := juiceconfig.GetInt("database.port")
boolValue, err := juiceconfig.GetBool("database.encrypted")
```

Default values can be provided.

```golang
stringValue, err := juiceconfig.GetString("database.hostname", "http://myDatabase.com")
intValue, err := juiceconfig.GetInt("database.port", 3306)
boolValue, err := juiceconfig.GetBool("database.encrypted", true)
```

If an error occurs, all subsequent calls will also fail. This allows you to perform multiple operations and ignore errors, then check for an error at the end.

```golang
hostname, _ := juiceconfig.GetString("database.hostname", "http://myDatabase.com")
port, _ := juiceconfig.GetInt("database.port", 3306)
isEncrypted, _ := juiceconfig.GetBool("database.encrypted", true)
if juiceconfig.WasError() {
  fmt.Printf("Fatal Error: Could not access application configuration: %s\n", juiceconfig.ErrorMessage())  
	fmt.Printf("Shutting down.\n")
  os.Exit(1)
}
```

The error status can be reset
```golang
juiceconfig.ResetError()
```

If you wish to load additional configurations, to from a location not specified by JUICE_CONFIG, use the Load function. The same functions and error checking functions can be used as described above.

```golang
myConfig := juiceconfig.ResetError("file:::/path/to/my/config/file.json")
stringValue, err := myConfig.GetString("variable.name")
```

## Non-JSON Config files

In some cases it is necessary to insert configuration variables into config files that are not accessed via JuiceConfig. For example, the Tomcat web server (Java) wants to access it's config files directly. For such situations, the NodeJS equivalent of this library can be used. See `node-juice-config` for details.
