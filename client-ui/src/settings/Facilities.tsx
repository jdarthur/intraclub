import * as React from 'react';
import {useGetFacilitiesQuery, useDeleteFacilityMutation} from "../redux/api.js";
import {Card, Empty, Space,} from "antd";
import {FacilityForm} from "./FacilityForm";
import {LabeledValue} from "../common/LabeledValue";
import {DeleteConfirm} from "../common/DeleteConfirm";
import {Ellipsis} from "../common/Ellipsis";

import {Facility} from "../model/facility"


export function Facilities() {
    const {data} = useGetFacilitiesQuery()

    const facilities = data?.resource?.map((f: Facility) => (
        <OneFacility key={f.id} id={f.id} user_id={f.user_id} address={f.address} name={f.name} courts={f.courts} layout_photo={f.layout_photo}/>
    ))

    return <div>
        {facilities?.length ? facilities : <Empty/>}
        <div style={{height: "1em"}}/>
        <FacilityForm/>
    </div>

}

export function OneFacility({id, user_id, name, address, courts, layout_photo}: Facility) {

    const [deleteFacility] = useDeleteFacilityMutation()

    const valueLink = <a href={`https://www.google.com/maps/place/${encodeURIComponent(address)}`} target={"_blank"}>
        {address}
    </a>

    const deleteSelf = () => {
        deleteFacility(id).then((res: { error: any, data: any }) => {
            if (res.error) {
                console.log(res.error)
            } else if (res.data) {
                console.log(res.error)
            }
        })
    }

    const initialState: Facility = {
        id, user_id, name, address, courts, layout_photo
    }

    const editForm = <FacilityForm FacilityId={id} InitialState={initialState} Update/>

    const title = <Ellipsis fullValue={name} maxLength={20}/>

    const extra = <Space>
        {editForm}
        <DeleteConfirm deleteFunction={deleteSelf} objectType={"facility"}/>
    </Space>

    return <Card title={title} style={{width: 200}} size={"small"}
                 extra={extra}>
        <LabeledValue label={"Address"} value={valueLink} vertical/>
        <LabeledValue label={"Number of courts"} value={courts} vertical/>
    </Card>
}