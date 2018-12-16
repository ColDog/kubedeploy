const http = require('http')
const port = 8080

const requestHandler = (req, res) => { res.end('Hello Node.js Server!') }

const server = http.createServer(requestHandler)
server.listen(port, (err) => {
  if (err) throw err;
  console.log(`server is listening on ${port}`)
})
