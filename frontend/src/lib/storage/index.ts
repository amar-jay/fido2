type ContainerStore = {[key:string]: any}
export class Store {
    private credentials: ContainerStore
    constructor() {
        this.credentials = {};
    }

    get(container:string, key:string): ContainerStore | any{
        const containerStore:ContainerStore = JSON.parse(localStorage.getItem(container) || "{}")
        this.credentials ||= containerStore

        return this.credentials[key]
    }

    getCredentials(key:string):ContainerStore {
        return this.get("credentials", key)
    }

    set(container:string, key:string, value:string) {
        this.get(container, key)

        this.credentials[key] = value

        return localStorage.setItem(container, JSON.stringify(this.credentials))
    }

    setCredentials(key:string, value:string) {
        return this.set("credentials", key, value)
    }
}

