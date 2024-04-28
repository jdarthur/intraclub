import {Button} from "antd";
import * as React from "react";
import {Link} from "react-router-dom";

type DebugAuthLinkArgs = {
    email: string,
    uuid: string,
    returnPath: string,
}

export function DebugAuthLink({email, uuid, returnPath}: DebugAuthLinkArgs) {

    let to = `/auth?email=${email}&uuid=${uuid}`
    if (returnPath) {
        to += "&return=" + returnPath
    }

    console.log(email, uuid, returnPath)

    return <Button style={{alignItems: "flex-end"}}>
        <Link to={to}>
            auth
        </Link>
    </Button>
}