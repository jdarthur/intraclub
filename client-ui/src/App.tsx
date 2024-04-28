import './App.css'
import * as React from 'react';
import {Login} from "./login/Login";
import {useToken} from "./redux/auth";

function App() {

    const token = useToken()

    return <div>
        {!token ? <Login/> : <div>no content</div>}
    </div>
}

export default App
