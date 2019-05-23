var Unibabel = require('unibabel').Unibabel

var getDatabase = require('./database')

module.exports = getEvents

// getEvents queries the server API for events using the given query parameters.
// Once the server has responded, it looks up the matching UserSecrets in the
// local database and decrypts and parses the previously encrypted event payloads.
function getEvents (query) {
  return window
    .fetch(`${process.env.SERVER_HOST}/events`, {
      method: 'GET',
      credentials: 'include'
    })
    .then(function (response) {
      if (response.status >= 400) {
        return response.json().then(function (errorBody) {
          var err = new Error(errorBody.error)
          err.status = response.status
          throw err
        })
      }
      return response.json()
    })
    .then(function (payload) {
      var db = getDatabase()
      var decrypted = payload.events.map(function (event) {
        return db.secrets.get({ accountId: event.account_id })
          .then(function (result) {
            var userSecret = result.userSecret
            return window.crypto.subtle
              .decrypt({
                name: 'AES-CTR',
                counter: new Uint8Array(16),
                length: 128
              }, userSecret, Unibabel.base64ToArr(event.payload))
              .then(function (decrypted) {
                var payloadAsString = Unibabel.utf8ArrToStr(new Uint8Array(decrypted))
                return Object.assign({}, event, { payload: JSON.parse(payloadAsString) })
              })
          })
      })
      return Promise.all(decrypted)
    })
}