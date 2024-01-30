/* note:
  *    1.  the following code is a demo to show how to build WebRTC P2P connection with GT client.
  *    2.  the value of stunUrl (line 9) 'stun:somebody.gtserver.com:3478' is a placeholder, you should replace it with the stun address of your own GT server
  *    3.  the value of gtServer(line 48) 'https://somebody.gtserver.com' is a placeholder, you should replace it with the tunnel address of your own GT server
 */
async function main() {
    //step one: define the variables of stun server
    // todo: this is a placeholder, you should replace it with the stun address of your own GT server
    const stunUrl = 'stun:somebody.gtserver.com:3478'

    //step two: create a RTCPeerConnection object
    const rtcConnection = new RTCPeerConnection({
        iceServers: [
            {
                urls: stunUrl
            }
        ]
    })

    //step three: create a variable to store the candidate and local description
    const resArray = []

    //step four: create a listener to get the candidate
    rtcConnection.onicecandidate = (event) => {
        if (event.candidate) {
            resArray.push(event.candidate)
        }
    }

    // step five: create offer and store local description
    rtcConnection
        .createOffer()
        .then((desc) => {
            rtcConnection.setLocalDescription(desc)
            resArray.push(desc)
        })
        .catch((err) => {
            console.error('Failed to get Offer!', err)
        })

    // step six: send the candidate and local description to the server
    // we wait for 5 seconds, if the result is not ready, we still send the result to the server
    await wait(5000)

    //step seven: convert the resArray to ArrayBuffer, and send it to the server
    const buffer = await convertResArrayToArraybuffer(resArray)
    // todo: this is a placeholder, you should replace it with the tunnel address of your own GT server
    const gtServer = 'https://somebody.gtserver.com'
    const response = await fetch(gtServer, {
        method: 'XP',
        body: buffer,
        headers: {
            'Transfer-Encoding': 'chunked'
        }
    })
    //step eight:get the response and convert it to remoteResArr
    const remoteDesAndCandidate = await response.arrayBuffer()
    const remoteResArr = parseBufferToResArray(remoteDesAndCandidate)

    //step nine:loop remoteResArr
    remoteResArr.forEach((item) => {
        try {
            let data = JSON.parse(item)
            if (data.type === 'offer') {
                rtcConnection.setRemoteDescription(new RTCSessionDescription(data))
            } else if (data.type == 'answer') {
                rtcConnection.setRemoteDescription(new RTCSessionDescription(data))
            } else if (data.candidate) {
                let candidate = new RTCIceCandidate({
                    sdpMLineIndex: data.sdpMLineIndex,
                    candidate: data.candidate
                })
                rtcConnection.addIceCandidate(candidate)
            }
        } catch (e) {
            console.log(e)
        }
    })

    // step ten: you can create datachannel now
    const datachannel = rtcConnection.createDataChannel('label-1')
    datachannel.onopen = () => {
    }
    datachannel.onmessage = (event) => {
        console.log(event.data)
    }
}


main()


function wait(times) {
    const {resolve, promise} = genPromise()
    setTimeout(() => {
        resolve('ok,times:' + times + ' is over')
    }, times)
    return promise
}

function genPromise() {
    let promise, resolve, reject
    promise = new Promise((r, j) => {
        resolve = r
        reject = j
    })
    return {promise, resolve, reject}
}

// todo:  you should convert the resArray to  Blob, the convert standard you can find  in  https://github.com/ao-space/gt
function convertResArrayToArraybuffer(resArray) {
    const blob = new Blob(resArray)
    // then convert the Blob to ArrayBuffer
    return blob.arrayBuffer()
}

// todo:  you should convert the buffer to string, the convert standard you can find  in https://github.com/ao-space/gt
function parseBufferToResArray(buffer) {
    let resArray = []
    return resArray
}