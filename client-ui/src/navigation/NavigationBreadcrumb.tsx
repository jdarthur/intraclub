import {Breadcrumb} from "antd";
import * as React from "react";
import {Link} from "react-router-dom";

type BreadcrumbItem = {
    title: React.ReactNode;
}

type BreadcrumbProps = {
    items: React.ReactNode[]
}

export function NavigationBreadcrumb({items}: BreadcrumbProps) {

    const homeLink: BreadcrumbItem = {
        title: <Link to={"/"}>Home</Link>
    }
    const breadcrumbItems = [homeLink]

    for (let item of items) {
        breadcrumbItems.push({title: item})
    }

    return <div style={{marginBottom: "1em"}}>
        <Breadcrumb items={breadcrumbItems}/>
    </div>
}