import {Button} from "antd";
import * as React from "react";
import {Link} from "react-router-dom";

type DebugAuthLinkArgs = {
    email: string,
    token: string,
    returnPath: string,
}

export function DebugAuthLink({email, token, returnPath}: DebugAuthLinkArgs) {

    let to = `/auth?token=${token}`
    if (returnPath) {
        to += "&return=" + returnPath
    }

    console.log(email, token, returnPath)

    return  <Link to={to}>
        <Button style={{width:'100%', marginTop:'10px'}}>
            auth
        </Button>
    </Link>
}