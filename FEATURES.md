# Features
- [ ] Backup database(s)
  - [x] Mysql Dump
    - [x] Filesystem
    - [x] Dropbox
  - [x] Bin logs
    - [x] Backup binlogs of a specific database of a specific backup
    - [x] Check the todos
    - [x] Clean current implementation to use (services, models & jobs)
  - [x] Add proper error handling when either method or storage fail
  - [x] Recovery
- [ ] Users & privileges management
  - [x] Create user
  - [x] List users
  - [ ] Validations for create user
  - [ ] Update user
  - [ ] Delete user
  - [ ] View user details(everything possible including privileges)
  - [ ] Add/Remove privileges on a specific user
- [ ] Configuration
- [ ] Monitoring using [prometheus](https://prometheus.io/)
- [ ] Add support for running as a service
- [ ] Deploy the first version ever of **Oversight**
- [ ] Replication

### CLI client
- [ ] All the features above

### Web client
- [ ] All the features in CLI
- [ ] MySql client

### Future
- [ ] Ovrsight mysql server <!-- Creating an ovrsight managed Mysql database server from scratch -->
- [ ] Horizontal scaling
- [ ] Percona Extra Backup as an additional backup method
- [ ] MyDumper as an additional backup method