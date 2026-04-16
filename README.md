

<div style="text-align: center;">
    <img src="resources/images/logo_with_name_with_bg.png" alt="logo" width="200"/>
</div>

# Vivactor-M

Vivactor-M is a specialized tool designed to facilitate real-time microservices refactoring with usage of liveness features (real-time feedback). It aims to improve architectural quality by providing immediate feedback and automated transformation.


## Requirements

- Docker

## 📥 Installation and Usage

1. **Insert your own API Key of Gemini at ./architecture-retrieval/.env, like this**:
```bash
GEMINI_API_KEY=blahblahblah...
```

1. **Clone the repository:**
```bash
git clone https://github.com/JocaFerna/Vivactor-M.git
```
2. **Start all the software**
```bash
./start_scripts/start_architecture.sh
```
Note: You may need to give permissions to execute this script, you can do it using this command

```bash
sudo chmod +x ./start_scripts/start_architecture.sh
```


3. **Access http://localhost:5173/**

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](https://www.google.com/search?q=LICENSE) file for details.

## Notes

- The only .env file on .gitignore is the one that contains the GEMINI_API_KEY (./architecture_retrieval/.env). This is on purpose, since none of the data (except the Gemini API Key) are private data and, this allows future implementations to easily coordinate data within the services.

- If, for some reason it states something like this when starting the architecture:

```bash
Error response from daemon: cannot remove container {SERVICE_NAME}: container is running: stop the container before removing or force remove
```

You can run the script that mitigates this issue:

```bash
./help_scripts/docker_clear_by_force.sh
```
Note: You may need to give permissions to execute this script, you can do it using this command
```bash
sudo chmod +x ./help_scripts/docker_clear_by_force.sh
```

## 👤 Author

**João Fernandes (JocaFerna)**

* GitHub: [@JocaFerna](https://www.google.com/search?q=https://github.com/JocaFerna)
* Project Link: [https://github.com/JocaFerna/LiveRefactoringTool](https://github.com/JocaFerna/LiveRefactoringTool)