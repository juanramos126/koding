fs     = require 'fs'
faker  = require 'faker'

module.exports =

  generateUsers: ->

    users = []

    for i in [1..1]

      name     = faker.Name.findName()
      username = faker.Helpers.slugify(faker.Internet.userName()).toLowerCase().replace(/\./g, '').replace(/_/g, '')
      username = username.substring(0, 7) + Date.now()
      password = @getPassword()
      teamSlug = name.toLowerCase().replace /\s/g, '-'

      email = "kodingtestuser+#{username}@koding.com"

      users.push { name, email, username, password, teamSlug }


    fs.writeFileSync 'users.json', JSON.stringify(users), 'utf-8'

    return users


  getPassword: ->

    password = faker.Helpers.slugify(faker.Internet.userName())

    while password.length < 12
      password = faker.Helpers.slugify(faker.Internet.userName())

    return password


  getUser: (createNewUserData) ->

    if createNewUserData
      users = @generateUsers()
      return users[0]

    try
      usersFile = fs.readFileSync('users.json')
      users = JSON.parse(usersFile)

      console.log ' ✔ users.json found, returning first user'
      return users[0]

    catch
      console.log ' ✔ users.json does not exist, creating new user data'

      users = @generateUsers()
      return users[0]
