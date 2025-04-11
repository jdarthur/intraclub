import * as React from 'react';
import {useCreateOneTimePasswordMutation} from "../redux/api.js";
import {Alert, Button, Card, Tooltip} from "antd";
import {useState} from "react";
import LabeledInput from "../common/LabeledInput";
import {InfoCircleOutlined, UserOutlined} from "@ant-design/icons";
import {DebugAuthLink} from "./DebugAuthLink";
import {useParams, useSearchParams} from "react-router-dom";

export function Login() {
    const [createOneTimePassword, {isLoading}] = useCreateOneTimePasswordMutation({});
    const [email, setEmail] = useState<string>("")
    const [error, setError] = useState<string>("")
    const [success, setSuccess] = useState<boolean>(false)
    const [token, setToken] = useState<string>("")

    const [searchParams, setSearchParams] = useSearchParams()
    const returnTo = searchParams.get('return')

    const login = async () => {

        const body = {
            "email": email
        }

        const response = await createOneTimePassword(body)
        if (response?.error) {
            setError(response?.error?.data?.error)
        } else if (response?.data) {
            setError("")
            setSuccess(true)
            setToken(response?.data?.token.token)
        }
    }

    const title = <div style={{width: "100%", display: "flex", justifyContent: "center"}}>
        <UserOutlined style={{marginRight: "0.8em"}}/>
        Log in
    </div>

    const buttonDisabled = isLoading || success == true || email == ""

    return <Card title={title} style={{width: 300}}>
        <div style={{display: "flex", flexDirection: "column", padding: "0.5em"}}>
            <LabeledInput label={"Email address"} value={email} setValue={setEmail} placeholder={"Enter email"}
                          style={{marginBottom: "1em"}} disabled={success} onEnter={login}/>
            {error ? <Alert type={"error"} message={error} style={{marginBottom: "1em"}}/> : null}
            {success ? <Alert type={"success"} message={LoginCodeMessage(email)} style={{marginBottom: "1em"}}/> : null}
            <Button onClick={login} type={"primary"} disabled={buttonDisabled} loading={isLoading}>Send login
                code</Button>
            {success ? <DebugAuthLink email={email} token={token} returnPath={returnTo}/> : null}
        </div>
    </Card>

}

function LoginCodeMessage(email: string): React.JSX.Element {
    return <span>
        A new login code has been sent to email <i>{email}</i>
        <Tooltip placement={"bottom"}
                 title={"Your login code should arrive from noreply@rcintra.club. Check your spam folder if you haven't received it in the next few minutes."}>
            <InfoCircleOutlined style={{marginLeft: "0.5em"}}/>
        </Tooltip>
</span>
}