import * as React from 'react';
import {Button, Form, Input, InputNumber, Modal} from "antd";
import {FormItem} from "../common/FormItem";
import {PlusSquareOutlined} from "@ant-design/icons";
import {useCreateFacilityMutation, useUpdateFacilityMutation} from "../redux/api.js" ;
import {Facility} from "./Facilities";

type FacilityFormProps = {
    Update?: boolean
    FacilityId?: string
    InitialState?: Facility
    button?: React.ReactNode
}

export function FacilityForm({Update, InitialState, FacilityId, button}: FacilityFormProps) {
    const [open, setOpen] = React.useState<boolean>(false);
    const [form] = Form.useForm()

    const [createFacility] = useCreateFacilityMutation()
    const [updateFacility] = useUpdateFacilityMutation()

    const onSubmit = (v: any) => {
        const values = form.getFieldsValue();

        const body: Facility = {
            name: values.name,
            address: values.address,
            courts: values.courts
        }

        if (Update) {
            updateFacility({id: FacilityId, body}).then((res: { error: any; data: any }) => {
                if (res.error) {
                    console.log(res.error)
                } else if (res.data) {
                    setOpen(false)
                }
            })
        } else {
            createFacility(body).then((res: { error: any; data: any }) => {
                if (res.error) {
                    console.log(res.error)
                } else if (res.data) {
                    setOpen(false)
                }
            })
        }
    }

    const DefaultButton = <Button type={"primary"}>
        <PlusSquareOutlined/>
        New Facility
    </Button>

    const actionButton = <div onClick={() => setOpen(true)}>
        {button ? button : DefaultButton}
    </div>
    return <div>
        <Modal open={open} title={"Create a new facility"} onCancel={() => setOpen(false)} onOk={onSubmit}>
            <div style={{height: "1em"}}/>
            <Form form={form}
                  style={{display: "flex", flexDirection: "column"}}
                  layout={"horizontal"}
                  initialValues={InitialState}>
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
        {actionButton}
    </div>
}