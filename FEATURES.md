# Features
- [x] Backup database(s)
  - [x] Mysql Dump
    - [x] Filesystem
    - [x] Dropbox
  - [x] Bin logs
    - [x] Backup binlogs of a specific database of a specific backup
    - [x] Check the todos
    - [x] Clean current implementation to use (services, models & jobs)
  - [x] Add proper error handling when either method or storage fail
  - [x] Recovery
- [x] Users & privileges management
  - [x] Create user
  - [x] List users
  - [x] Validations for create user
  - [x] Update user
  - [x] Delete user
  - [x] View user details(everything possible including privileges)
  - [x] Add/Remove privileges on a specific user
- [x] Configuration
- [ ] Improve backup and restoration
- [ ] Add support for setting up monitoring using [grafana](https://grafana.com/grafana/dashboards/)
- [ ] Enable/disable Replication
- [ ] Enable/disable Horizontal scaling

### Web client
- [ ] All the features in CLI
- [ ] API