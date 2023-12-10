import { useState, useEffect } from 'react'

/**
 *
 * implemented in https://github.com/teamhanko/hanko/blob/1af3958afb108fbaab205d9cabf5e3da47b9dec3/frontend/frontend-sdk/src/lib/client/WebauthnClient.ts#L96
import webauthn, {
  create as createWebauthnCredential,
  get as getWebauthnCredential,
} from "@github/webauthn-json";
 */

import {Store} from "./lib/storage"
import './App.css'


type AssertionResult = {
  credential_id: string,
  user_id: string
  }

async function loginFido(store:Store, user_id:string){
    // initialize
    const req = await fetch("http://localhost:8000/register/"+user_id + "/begin",)
    // TODO: proper handling of request 
    
    const res:CredentialCreationOptions = await req.json()!

    const assertion = await navigator.credentials.get(res)
    if (!assertion){
      throw new Error("assertion is null")
      }

    // finilize 
    const assertionReq = await fetch("http://localhost:8000/register/"+ user_id + "/end",{
      method: "POST",
      body: JSON.stringify(assertion)
      })
    const finalRes:AssertionResult = await assertionReq.json()!

    store.setCredentials(user_id, finalRes.credential_id)
    return {
      finalRes,
      assertion
      }
}


function wait(ms:number) {
  return new Promise(resolve => setTimeout(resolve, ms));
}
async function registerFido(store:Store, user_id:string) {
    // Check for WebAuthn support
    const res = await fetch("http://localhost:8000/register/"+user_id + "/begin")
    // TODO: proper handling of request 

     const response:CredentialCreationOptions = await res.json()!

     if (!response.publicKey){
       throw new Error("public key error")
       }
     response.publicKey!.timeout = 10000
     response.publicKey!.challenge = new ArrayBuffer(response.publicKey!.challenge as any)
     response.publicKey!.user.id = new ArrayBuffer(response.publicKey!.user.id as any)
     const userId:string = response.publicKey.user.id as any

     const credentials = await navigator.credentials.create(response) // has no attestation

     
     const fetchRes = await wait(10000)
     .then(() => fetch("/webauthn/registration/"+user_id+"/end",{
         body: JSON.stringify(credentials)
       }))
       .then(e => e.json())
       .then(finalRes => {
        store.setCredentials(userId, finalRes.credential_id)
        return finalRes
        })


     return {
       credentials,
       fetchRes
       }
}

/**
         let assertion;
    try {
      assertion = await getWebauthnCredential(response.publicKey.challenge);
    } catch (e) {
      throw new Error("WebAuthn Challenge error");
    }

    const assertionReq = await fetch("http://localhost:8000/authenticate/"+assertion)
    const assertionRes = await assertionReq.json()
     return cred
 */

function App() {
  const store = new Store()
  const [user, setUser] = useState("user")
  const [cred, setCred] = useState<Credential|null>(null)
  const [err, setError] = useState<string>("")

  useEffect(() => {
    if (!window.PublicKeyCredential) {
      setError("WebAuthn not supported in this browser.")
    }
    }, [])

  const register = async (user:string) => {

        await registerFido(store,user)
        .then(({credentials}) => { 
          if (cred) {
            setCred(credentials)
          } else {
            let e = "credentials is null"
            console.error(e)
            setError(e)
            } 
        })
        .catch((e) => {
        let error = e as Error
        console.error("Error creating credential:", error);
        setError(error.name + ": "+ error.message)
        })

  }

  const login = async (user_id:string) => {
    try{
        const {assertion} = await loginFido(store,user)
        console.log(assertion, user_id)
    } catch (e) {
        let error = e as Error
        console.error("Error creating credential:", error);
        setError(error.name + ": "+ error.message)
      }
  }
  return (
    <div className="App">
    <input className="input" onChange={e=> setUser(e.target.value)} value={user}/>
    <div>
      <div>
      <h3> Using the default </h3>
          <p>
          {cred ? (cred?.id + " " + cred?.type): "void credentials"}
          </p>
          <button onClick={() => register(user)} className='dark'> Register </button>
          <div/>
          <button onClick={() => login(user)}> Login </button>
      </div>

      <div>
      <h3> Using the @github/webauth-json </h3>
          <p>
          not implemented yet
          </p>
      </div>
      </div>

          {
            err ? (<h4 className="red">{JSON.stringify(err)}</h4>): null
          }
    </div>
  )
}

export default App
