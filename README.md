# ASO - Arnold's Super Organizer

ASO (Arnold's Super Organizer) is a powerful program designed to help you efficiently manage and organize your students
or user groups within your GitHub projects. Whether you're an educator managing a class project or an organization with
multiple collaborators, ASO simplifies group management and enhances your workflow.

## 👾 Features

- **Create User Groups**: ASO allows you to create user groups for your GitHub project. Each user group can have an
  optional expiration date, ensuring that users are automatically removed from the project when the specified date is
  reached.

- **User Group Actions**: Perform various actions on user groups, including adding all users to the Git repository or
  removing them from the repository beforehand.

- **QR Code Integration**: ASO makes it incredibly easy to add users to your user groups. It generates a QR code that
  users can scan to join a specified user group, automatically granting them access to the repository.

- **Effortless Setup**: ASO is designed with user-friendliness in mind. Setting up and using the program is a breeze,
  thanks to clear and intuitive instructions.

- **Automatic RSA Key Generation**: ASO automatically generates an RSA key for jwt authentication, saving you the hassle
  of generating one yourself.\
  If you already have an RSA key, you can simply replace the existing key with your own under `./rsa_private_key.pem`

## 🚀 Getting Started

To run ASO, make sure you have the following prerequisites:

- MongoDB installed and running.
- Go programming language installed.
- An `.env` file in the directory of the source code, with the following environment variables:

```shell
MONGODB_URI=mongodb://localhost:27017
MONGODB_DB=ASO
PORT=<server port> (optional default port: 8080)
```

## 💻 Automatic Setup

1. **Clone the Repository**:<br>
   Begin by cloning the ASO repository to your local machine.
   ```shell
   git clone https://github.com/MindCollaps/ASO
   ```
2. **Run the install Script**:<br>
   Execute the install.sh script to install ASOServer as a systemd service.
   ```shell
   sudo chmod +x ./install.sh && sudo ./install.sh
   ```
   This script can also update your service if you have already installed it.

## 🛠️ Manual Setup

1. **Clone the Repository**:<br>
   Begin by cloning the ASO repository to your local machine.
   ```shell
   git clone https://github.com/MindCollaps/ASO
   ```

2. **Set Up Environment Variables**:<br>
   Create an .env file in the root directory and add your MongoDB URI and the optional
   PORT and DB, as mentioned in the prerequisites.

3. **Compile the Program**:<br>
   Build the program using the following command in the ASO directory:
   ```shell
   go build
   ```

4. **Run the Program**:<br>
   Execute the program using the following command:
   ```shell
   ./ASOServer
   ```

## 🌈 MISC

### 📜 Install Script

The `install.sh` script offers a streamlined solution for installing ASOServer on Unix-based systems as a systemd
service.

### Command Line Arguments (install.sh)

- ```install.sh --go <path>```: Change go path

### 🔎 Where to find stuff

- **Binary** : ```/usr/local/sbin/ASOServer```
- **Config and Keys** : ```/etc/aso/```
- **Logs** : ```/var/log/aso/aso.log```

## 🧑‍⚖️ License

ASO is an open-source project released under the MIT License.

We hope you find ASO helpful in organizing your GitHub projects. If you have any questions or encounter issues, please
don't hesitate to reach out to us. Happy organizing! 💼🚀👩‍🏫👨‍💻📚
