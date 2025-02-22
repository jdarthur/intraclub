import * as React from 'react';
import {Form} from "antd";
import {InputFormItem, NumberInputFormItem} from "../common/FormItem";
import {useCreateFacilityMutation, useUpdateFacilityMutation} from "../redux/api.js" ;
import {Facility} from "./Facilities";
import {CommonFormModal} from "../common/CommonFormModal";

type FacilityFormProps = {
    Update?: boolean // is this updating an existing record or creating a new record
    FacilityId?: string // this will be provided on an update
    InitialState?: Facility // this will be provided on an update
}

export function FacilityForm({Update, InitialState, FacilityId,}: FacilityFormProps) {

    const [createFacility] = useCreateFacilityMutation()
    const [updateFacility] = useUpdateFacilityMutation()

    const onSubmit = async (formValues: any) => {
        const body: Facility = {
            name: formValues.name,
            address: formValues.address,
            courts: formValues.courts
        }

        let func = () => createFacility(body)
        if (Update) {
            func = () => updateFacility({id: FacilityId, body: body})
        }

        return await func();
    }


    return <CommonFormModal ObjectType={"facility"} IsUpdate={Update} OnSubmit={onSubmit} InitialState={InitialState}>
        <InputFormItem name={"name"} label={"Name"}/>
        <InputFormItem name={"address"} label={"Address"}/>
        <NumberInputFormItem name={"courts"} label={"Number of courts"}/>
    </CommonFormModal>
}
