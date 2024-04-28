import {Layout} from "antd";
import * as React from "react";
import {NavMenu} from "./toolbar";

const {Header, Content} = Layout;

type MainLayoutProps = {
    content: React.ReactNode;
}

export function MainLayout({content}: MainLayoutProps) {
    return (
        <Layout style={{width: '100vw', height: "100vh"}}>
            <Header>
                <NavMenu/>
            </Header>
            <Content style={{padding: '1em'}}>
                {content}
            </Content>
        </Layout>
    )
}