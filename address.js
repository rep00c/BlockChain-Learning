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

let privateKey = createRandomPrivateKey()
let address = privateKeyToAddress(privateKey)
console.log('0x'+address.toString('hex'))
console.log(privateKey.toString('hex'))