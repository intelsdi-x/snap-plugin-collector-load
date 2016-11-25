# Example tasks

[This](task-load.json) example task will collect metrics from **load** and publish
them to file.  

## Running the example

### Requirements 
 * `docker` and `docker-compose` are **installed** and **configured** 

Running the sample is as *easy* as running the script `./run-mock-load.sh`.

## Files
- [mock-load.sh](mock-load.sh)
    - Downloads `snapteld`, `snaptel`, `snap-plugin-collector-load`,
        `snap-plugin-publisher-mock-file` and starts the task
- [run-mock-load.sh](run-mock-load.sh)
    - The example is launched with this script     
- [task-load.json](task-load.json)
    - Snap task definition
- [.setup.sh](.setup.sh)
    - Verifies dependencies and starts the containers.  It's called 
    by [run-mock-load.sh](run-mock-load.sh).
- [docker-compose.yml](docker-compose.yml)
    - A docker compose file which defines "runner" container where snapteld
     is run from. You will be dumped into a shell in this container
     after running.    