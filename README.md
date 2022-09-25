# DisCache: Distributed Cache

What is the purpose of Discache?

Web applications require a way to store data into db in a fast way. They usually write to a SQL database (Postgres, MySQL, etc.) and then wait
for the data to be written/committed to continue their work. One of the problems with this method is that performance depends on the disk speed
and also depends on the number of write requests.

Discache is a distributed cache server which stores data into a tmp cache and then push it into SQL/NoSQL database without losing data.

* The "tmp cache" is reliable due to cache replication. At least 2 instances of the cache should be available at any time.
* Discache implements simple get/put functionality and also a "distributed sorted-set" similar to redis sorted-set.
* Discache may store json data into corresponding SQL tables 



### To implement

* Implement a cache package store key/value data
* Implement a dCache package which is responsible for storing/retrieving data from every cache server. It handles the distribution.
* Implement a probe package which constantly check the status (healthy/broken statuses) of cache servers. 
* Implement a go-client which uses http protocol to push data into the discache server
* Implement a worker responsible for inserting data from queue to SQL tables
* test coverage
* Dockerfile which runs the server
* "docker-compose.yaml" which starts X nodes for local testing
* K8S deployment