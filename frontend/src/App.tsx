import { useState, useEffect } from 'react'
import './App.css'

function authenticateFido() {
  }
async function registerFido(user:string) {
    // Check for WebAuthn support
    const res = await fetch("http://localhost:8000/register/"+user,)

     const response = await res.json()
     console.log(response)
     
     const cred = await navigator.credentials.create({
       publicKey: {
         ...response.publicKey,
         timeout: 30000000,
         attestation: "none",
         challenge: new ArrayBuffer(response.publicKey.challenge),
         user: {
           ...response.publicKey.user,
           id: new ArrayBuffer(response.publicKey.user.id)
          }
          
         },

       })
      console.log(cred)
      return cred
  }
function App() {
  const [user, setUser] = useState("user")
  const [cred, setCred] = useState<Credential|null>(null)
  const [err, setError] = useState<string>("")

  useEffect(() => {
    if (!window.PublicKeyCredential) {
      setError("WebAuthn not supported in this browser.")
    }
    }, [])

  const register = async (user:string) => {

    try{
        const cred = await registerFido(user)
        if (cred) {
          setCred(cred)
        } else {
          let e = "credentials is null"
          console.error(e)
          setError(e)
          } 
      } catch(e){
        let error = e as Error
        console.error("Error creating credential:", error);
        setError(error.name + "\n"+ error.message)
        }

  }
  return (
    <div className="App">
    <input onChange={e=> setUser(e.target.value)} value={user}/>
      <p>
      {cred ? (cred?.id + " " + cred?.type): "void credentials"}
      </p>
      <button onClick={() => register(user)} className='dark'> Register </button>
      <div/>
      <button onClick={authenticateFido}> Authenticate </button>
      {
        err ? (<h4>{JSON.stringify(err)}</h4>): null
      }
    </div>
  )
}

export default App
