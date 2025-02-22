import * as React from 'react';
import {useRegisterMutation} from "../redux/api.js";
import {Alert, Button, Card, Form, Space, Tooltip} from "antd";
import {useState} from "react";
import {InfoCircleOutlined, UserAddOutlined, UserOutlined} from "@ant-design/icons";
import {InputFormItem} from "../common/FormItem";

type RegisterBody = {
    email: string
    first_name: string
    last_name: string
}

export function Register() {
    const [sendRegister, {isLoading}] = useRegisterMutation();

    const [values, setValues] = useState<RegisterBody>({
        email: undefined, first_name: undefined, last_name: undefined
    })
    const [error, setError] = useState<string>("")
    const [success, setSuccess] = useState<boolean>(false)

    const [form] = Form.useForm()

    const register = async () => {
        const response = await sendRegister(values)
        if (response?.error) {
            setError(response?.error?.data?.error)
        } else if (response?.data) {
            setError("")
            setSuccess(true)
        }
    }

    const onChange = () => {

        const v = form.getFieldsValue()

        const newFormValues: RegisterBody = {
            email: v.email,
            first_name: v.first_name,
            last_name: v.last_name
        }

        setValues(newFormValues)
    }

    const title = <div style={{display: "flex", justifyContent: "center"}}>
        <UserAddOutlined style={{marginRight: "0.8em"}}/>
        Create a new account
    </div>

    const buttonDisabled = success || isLoading || !values.email || !values.first_name || !values.last_name

    return <Card title={title} style={{width: 400}}>
        <div style={{display: "flex", flexDirection: "column", padding: "0.5em"}}>
            <Form form={form} onChange={onChange}>
                <InputFormItem name={"email"} label={"Email"} disabled={success}/>
                <InputFormItem name={"first_name"} label={"First name"} disabled={success}/>
                <InputFormItem name={"last_name"} label={"Last name"} disabled={success}/>
            </Form>

            {error ? <Alert type={"error"} message={error} style={{marginBottom: "1em"}}/> : null}
            {success ?
                <Alert type={"success"} message={LoginCodeMessage(values.email)} style={{marginBottom: "1em"}}/> : null}
            <Button onClick={register} type={"primary"} disabled={buttonDisabled} loading={isLoading}>
                Register
            </Button>
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