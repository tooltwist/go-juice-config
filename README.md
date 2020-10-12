# go-juice-config

These libraries allow a program to seamlessly switch between configuration sources (e.g. Flat Files, and Amazon Secrets Manager).

Managing configurations as an application passes from development through to production can pose challenges.

- In a development environment, you want fast, easy-to-change access to your configuration information (e.g. In a JSON file).

- In production environments you need access to configuration stored securely (e.g. AWS Secrets Manager).

- If you are using Docker, you'll want to externalise the config from your application's Docker image, so the same image can be used in multiple environments (CI, test, UAT, Staging, Production, etc).

- Environment variables are too easy for hackers to access.


To solve these problems, JuiceConfig can be used by an application to seamlessly switch where config information is stored. An environment variable (JUICE_CONFIG) specifies where to find the configuration (but does not itself provide access to the configuration resources).


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
  
### Usage

