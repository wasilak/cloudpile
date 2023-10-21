import { useState, useEffect, useRef, useCallback } from "react";

import Box from '@mui/material/Box';
import CircularProgress from '@mui/material/CircularProgress';
import Stack from "@mui/material/Stack";
import Chip from '@mui/material/Chip';
// import List from '@mui/material/List';
// import ListItem from '@mui/material/ListItem';
// import ListItemText from '@mui/material/ListItemText';

import { AWSResource, ItemTag } from "./models";
import { GetAWSResources } from "./api";

import { AgGridReact } from "ag-grid-react";

import "ag-grid-community/styles/ag-grid.css";
import "ag-grid-community/styles/ag-theme-material.css";

import { ICellRendererParams, IRowNode } from 'ag-grid-community';

import Tooltip from '@mui/material/Tooltip';

import { FilterSelect } from './table/filterSelect'

export const ResourcesTable = () => {
    const [resources, setAWSResources] = useState<Array<AWSResource>>(undefined);
    const [rows, setRows] = useState([]);
    const [columns, setColumns] = useState([]);

    const gridRef = useRef<AgGridReact<AWSResource>>(null);

    // const TagsCellRenderer = (props: ICellRendererParams) => {
    //     return (
    //         <List dense={true}>
    //             {props.value.map((item: ItemTag) => {
    //                 return (
    //                     <ListItem key={`${item.key}=${item.value}`}>
    //                         <ListItemText
    //                             primary={`${item.key} = ${item.value}`}
    //                         />
    //                     </ListItem>
    //                 )
    //             })
    //             }
    //         </List>
    //     );
    // };

    const TagsCellRenderer = (props: ICellRendererParams) => {
        return (
            <Box>
                {props.value.map((item: ItemTag) => {
                    return (
                        <Box key={`${item.key}=${item.value}`}>
                            <Tooltip title={`${item.key}=${item.value}`}>
                                <Chip
                                    label={`${item.key}=${item.value}`}
                                    color="primary"
                                    variant="outlined"
                                />
                            </Tooltip>
                        </Box>
                    )
                })
                }
            </Box>
        );
    };

    const setupGridData = () => {
        if (resources) {

            const columns = [
                { field: 'id', sortable: true, filter: 'agTextColumnFilter', resizable: true, headerName: 'ID' },
                { field: 'type', sortable: true, filter: 'agTextColumnFilter', resizable: true, headerName: 'Type' },
                { field: 'account', sortable: true, filter: 'agTextColumnFilter', resizable: true, headerName: 'Account' },
                { field: 'accountAlias', sortable: true, filter: 'agTextColumnFilter', resizable: true, headerName: 'Account Alias' },
                { field: 'region', sortable: true, filter: 'agTextColumnFilter', resizable: true, headerName: 'Region' },
                { field: 'tags', sortable: true, resizable: true, headerName: 'Tags', cellDataType: 'object', cellRenderer: TagsCellRenderer, autoHeight: true, wrapText: false },
                { field: 'private_dns_name', sortable: true, filter: 'agTextColumnFilter', resizable: true, headerName: 'Private DNS' },
                { field: 'ip', sortable: true, filter: 'agTextColumnFilter', resizable: true, headerName: 'Private IP' },
                { field: 'arn', sortable: true, filter: 'agTextColumnFilter', resizable: true, headerName: 'ARN' },
            ];

            setColumns(columns);
            setRows(resources);

            const uniqueResourceTypes = Array.from(new Set(resources.map((item: AWSResource) => item.type)));

            setResourceTypes(uniqueResourceTypes);
        }
    };

    const [resourceTypes, setResourceTypes] = useState([]);
    const [resourceTypeSelected, setResourceTypeSelected] = useState('');

    const onChangeResourceTypes = (selectedOption: any) => {
        setResourceTypeSelected(selectedOption.target.value);
    }

    const isExternalFilterPresent = useCallback((): boolean => {
        console.log("isExternalFilterPresent", resourceTypeSelected !== '');
        return resourceTypeSelected !== '';
    }, []);

    const doesExternalFilterPass = useCallback(
        (node: IRowNode<AWSResource>): boolean => {
            console.log("doesExternalFilterPass", doesExternalFilterPass);
            if (node.data) {
                console.log(node.data);
                return node.data.type == resourceTypeSelected;
            }
            return true;
        },
        [resourceTypeSelected]
    );

    useEffect(() => {
        GetAWSResources(setAWSResources)
    }, []);

    useEffect(() => {
        setupGridData();
    }, [resources]);

    useEffect(() => {
        gridRef?.current?.api?.onFilterChanged();
    }, [resourceTypeSelected]);

    return (
        <Box sx={{ mt: 4 }}>
            {resources
                ? <Box >
                    <Box >
                        <FilterSelect items={resourceTypes} onChange={onChangeResourceTypes} selected={resourceTypeSelected} label="Type"></FilterSelect>
                    </Box>

                    <Box className="ag-theme-material" height="80vh" sx={{ mt: 4 }}>
                        <AgGridReact<AWSResource>
                            ref={gridRef}
                            rowData={rows}
                            columnDefs={columns}
                            pagination={true}
                            paginationAutoPageSize={true}
                            isExternalFilterPresent={isExternalFilterPresent}
                            doesExternalFilterPass={doesExternalFilterPass}
                        />
                    </Box>
                </Box>
                : <Box>
                    <Stack alignItems="center" >
                        <CircularProgress />
                    </Stack>
                </Box>
            }
        </Box>
    );
}
