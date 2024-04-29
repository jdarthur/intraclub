import * as React from 'react';
import {Button, Form, Input, InputNumber, Modal} from "antd";
import {FormItem} from "../common/FormItem";
import {PlusSquareOutlined} from "@ant-design/icons";
import {useCreateFacilityMutation} from "../redux/api";
import {Facility} from "./Facilities";

export function NewFacility() {
    const [open, setOpen] = React.useState<boolean>(false);
    const [form] = Form.useForm()

    const [createFacility] = useCreateFacilityMutation()

    const onSubmit = (v: any) => {
        const values = form.getFieldsValue();
        console.log(values)

        const body: Facility = {
            name: values.name,
            address: values.address,
            courts: values.courts
        }

        createFacility(body).then((res: { error: any; data: any }) => {
            if (res.error) {
                console.log(res.error)
            } else if (res.data) {
                setOpen(false)
            }
        })

    }

    return <div>
        <Modal open={open} title={"Create a new facility"} onCancel={() => setOpen(false)} onOk={onSubmit}>
            <div style={{height: "1em"}}/>
            <Form form={form} style={{display: "flex", flexDirection: "column"}} layout={"horizontal"}>
                <FormItem name={"name"} label={"Name"}>
                    <Input/>
                </FormItem>
                <FormItem name={"address"} label={"Address"}>
                    <Input/>
                </FormItem>
                <FormItem name={"courts"} label={"Number of courts"}>
                    <InputNumber/>
                </FormItem>
            </Form>
        </Modal>
        <Button onClick={() => setOpen(true)} type={"primary"}>
            <PlusSquareOutlined/>
            New Facility
        </Button>
    </div>
}