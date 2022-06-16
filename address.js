'use strict'

const randomBytes = require('randombytes')
const secp256k1 = require('secp256k1')
const keccak = require('keccak')

const createRandomPrivateKey = function () {
    return randomBytes(32)
}

const privateKeyToAddress = function (privateKey) {
    return keccak('keccak256').update(Buffer.from(secp256k1.publicKeyCreate(privateKey, false).slice(1))).digest().slice(-20)
}

const num = 10
const pri_arr = []
const pub_arr = []

for (let i = 0; i < num; i++) {
    const privateKey = createRandomPrivateKey()
    const address = privateKeyToAddress(privateKey)

    pub_arr.push('0x'+address.toString('hex'))
    pri_arr.push(privateKey.toString('hex'))
}

console.log(pub_arr, pri_arr)
