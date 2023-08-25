# DisCache: Distributed Cache

What is the purpose of Discache?

Web applications require a way to store data into db in a fast way. They usually write to a SQL database (Postgres, MySQL, etc.) and then wait
for the data to be written/committed to continue their work. One of the problems with this method is that performance depends on the disk speed
and also depends on the number of write requests.

Discache is a distributed cache server which stores data into a tmp cache and then push it into SQL/NoSQL database without losing data.

* The "tmp cache" is reliable due to cache replication. At least 2 instances of the cache should be available at any time.
* Discache implements simple get/put functionality and also a "distributed sorted-set" similar to redis sorted-set.
* Discache may store json data into corresponding SQL tables 



### FIXES

* Fix cluster name conflict. It's not a good idea to rely on "unique names" specified in conf file.


### TODO

* Add a conf template
* Implement a validation+default interface for inner configs  
* Add an entry to conf.Config for logging. Configure Zap to log into a file.