# Features
- [ ] Backup database(s)
  - [x] 1 db -> 1 storage: huge memory footprint
  - [x] 1 db -> 1 storage: small memory footprint
    - [ ] big file in chunks for Google Drive with one process
    - [ ] big file in chunks for Google Drive with multiple processes
    - [ ] add support for uploading chunks on Google Drive
  - [ ] bin logs
  - [ ] 1 db -> multiple backup methods
    - [ ] my_dumper with tests
    - [ ] percona extra_backup with tests
  - [ ] 1 db -> multiple storage at the same time
    - [ ] tests
  - [ ] multiple db -> multiple storage at the same time
    - [ ] tests
- [ ] Recovery
- [ ] Users & privileges management
- [ ] Configuration
- [ ] Ovrsight mysql server <!-- Creating an ovrsight managed Mysql database server -->
- [ ] Monitoring
- [ ] Replication
- [ ] Horizontal scaling

### CLI client
- [ ] All the features above

### Web client
- [ ] All the features in CLI
- [ ] MySql client