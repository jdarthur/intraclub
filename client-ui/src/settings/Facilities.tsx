import * as React from 'react';
import {useGetFacilitiesQuery, useDeleteFacilityMutation} from "../redux/api.js";
import {Card, Empty, Tag} from "antd";
import {NewFacility} from "./NewFacility";
import {LabeledValue} from "../common/LabeledValue";
import {DeleteConfirm} from "../common/DeleteConfirm";
import {Ellipsis} from "../common/Ellipsis";

export type Facility = {
    id?: string
    user_id?: string
    address: string
    name: string
    courts: number
    layout_image?: string
}

export function Facilities() {

    const {data} = useGetFacilitiesQuery()
    console.log(data)

    const facilities = data?.resource?.map((f: Facility) => (
        <OneFacility key={f.id} id={f.id} user_id={f.user_id} address={f.address} name={f.name} courts={f.courts}/>
    ))

    return <Card title={"Facilities"} style={{width: 'max(400, 95vw)'}}>
        {facilities?.length ? facilities : <Empty/>}
        <div style={{height: "1em"}}/>
        <NewFacility/>
    </Card>

}

export function OneFacility({id, name, address, courts, layout_image}: Facility) {

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

    const title = <Ellipsis fullValue={name} maxLength={20}/>

    return <Card title={title} style={{width: 200}} size={"small"}
                 extra={<DeleteConfirm deleteFunction={deleteSelf} objectType={"facility"}/>}>
        <LabeledValue label={"Address"} value={valueLink} vertical/>
        <LabeledValue label={"Number of courts"} value={courts} vertical/>
    </Card>
}