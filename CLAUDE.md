# Basic price oracle AVS

This is an EigenCloud Devkit project.

The main business logic of the AVS is located at cmd/main.go, and the main basic test is located at cmd/main_test.go.

For this AVS to function, the HandleTask function in main.go should always return a TaskId and a Result in the type of bytes []. 

When I ask you to make change to this project, never rerun the actual command line commands of the container or any of the tasks, only change the code, as I will be managing the rebuilding and redeployment of everything whenever there is a change.

