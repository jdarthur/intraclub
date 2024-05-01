import * as React from 'react';
import {Form} from "antd";
import {InputFormItem, NumberInputFormItem} from "../common/FormItem";
import {useCreateFacilityMutation, useUpdateFacilityMutation} from "../redux/api.js" ;
import {Facility} from "./Facilities";
import {CommonModal} from "../common/CommonModal";

type FacilityFormProps = {
    Update?: boolean // is this updating an existing record or creating a new record
    FacilityId?: string // this will be provided on an update
    InitialState?: Facility // this will be provided on an update
    button?: React.ReactNode // this is the button that you click to open the modal
}

export function FacilityForm({Update, InitialState, FacilityId, button}: FacilityFormProps) {

    const [form] = Form.useForm()

    const [createFacility] = useCreateFacilityMutation()
    const [updateFacility] = useUpdateFacilityMutation()

    const onSubmit = async () => {
        const values = form.getFieldsValue();
        const body: Facility = {
            name: values.name,
            address: values.address,
            courts: values.courts
        }

        let func = () => createFacility(body)
        if (Update) {
            func = () => updateFacility({id: FacilityId, body: body})
        }

        return await func();
    }

    return <CommonModal ObjectType={"facility"} IsUpdate={Update} OnSubmit={onSubmit}>
        <Form form={form} layout={"horizontal"} initialValues={InitialState}>
            <InputFormItem name={"name"} label={"Name"}/>
            <InputFormItem name={"address"} label={"Address"}/>
            <NumberInputFormItem name={"courts"} label={"Number of courts"}/>
        </Form>
    </CommonModal>
}


type FacilitySubmitArgs = {
    body: Facility
    Update: boolean
    FacilityId: string
    createFacility: ({}) => any
    updateFacility: ({}) => any
}

async function facilitySubmit({
                                  body,
                                  Update,
                                  FacilityId,
                                  updateFacility,
                                  createFacility
                              }: FacilitySubmitArgs) {


}