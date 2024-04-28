import * as React from 'react';
import {useToken} from "../redux/auth.js";
import {Button, Form, Input, Modal} from "antd";
import {PlusSquareOutlined} from "@ant-design/icons";

export type League = {
    LeagueId?
}

export function NewLeague() {

    const [open, setOpen] = React.useState<boolean>(false);

    const [form] = Form.useForm()

    const token = useToken()

    const onSave = (v: any) => {
        const formValues = form.getFieldsValue();
        console.log(formValues)
    }

    return <div>
        <Modal open={open} title={"Create a new league"} onCancel={() => setOpen(false)}>
            <Form form={form}
                  onFinish={onSave}>
                <Form.Item label={"Name"}>
                    <Input placeholder={"League name"}/>
                </Form.Item>
                <Form.Item label={"Name"}>
                    <Input placeholder={"League name"}/>
                </Form.Item>
            </Form>
        </Modal>
        {token ? <Button style={{marginTop: "2em"}} type={"primary"} onClick={() => setOpen(true)}>
            <PlusSquareOutlined style={{marginRight: "0.5em"}}/>
            Create a new league
        </Button> : null}
    </div>
}