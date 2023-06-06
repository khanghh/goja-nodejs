const sub = require('./submodule')

function test() {
  return sub.value;
}

module.exports = {
    test: test
}
