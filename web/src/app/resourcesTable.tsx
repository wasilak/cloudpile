import { useState, useEffect, useRef, useCallback } from "react";

import Box from '@mui/material/Box';
import CircularProgress from '@mui/material/CircularProgress';
import Stack from "@mui/material/Stack";
import Chip from '@mui/material/Chip';

import { AWSResource, ItemTag } from "./models";
import { GetAWSResources } from "./api";

import { AgGridReact } from "ag-grid-react";

import "ag-grid-community/styles/ag-grid.css";
import "ag-grid-community/styles/ag-theme-material.css";

import { ICellRendererParams, IRowNode } from 'ag-grid-community';

import Tooltip from '@mui/material/Tooltip';

import { FilterSelect } from './table/filterSelect'

import Grid from '@mui/material/Unstable_Grid2';

export const ResourcesTable = () => {
    const [resources, setAWSResources] = useState<Array<AWSResource>>(undefined);
    const [rows, setRows] = useState([]);
    const [columns, setColumns] = useState([]);

    const gridRef = useRef<AgGridReact<AWSResource>>(null);

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

            const uniqueRegions = Array.from(new Set(resources.map((item: AWSResource) => item.region)));
            setRegions(uniqueRegions);

            const uniqueAccounts = Array.from(new Set(resources.map((item: AWSResource) => `${item.accountAlias} (${item.account})`)));
            setAccounts(uniqueAccounts);

        }
    };

    const [resourceTypes, setResourceTypes] = useState([]);
    const [type, setResourceTypeSelected] = useState('');

    const onChangeResourceTypes = (selectedOption: any) => {
        setResourceTypeSelected(selectedOption.target.value);
    }

    const [regions, setRegions] = useState([]);
    const [region, setRegionSelected] = useState('');

    const onChangeRegions = (selectedOption: any) => {
        setRegionSelected(selectedOption.target.value);
    }

    const [accounts, setAccounts] = useState([]);
    const [account, setAccountSelected] = useState('');

    const onChangeAccounts = (selectedOption: any) => {
        setAccountSelected(selectedOption.target.value);
    }

    const isExternalFilterPresent = (): boolean => {
        return type !== '' || region !== '';
    };

    const doesExternalFilterPass = useCallback(
        (node: IRowNode<AWSResource>): boolean => {
            let outcome = true;

            if (node.data) {

                if (region != '') {
                    if (outcome && node.data.region == region) {
                        outcome = true;
                    } else {
                        return false;
                    }
                }

                if (account != '') {
                    if (outcome && node.data.account == account) {
                        outcome = true;
                    } else {
                        return false;
                    }
                }

                if (type != '') {

                    if (outcome && node.data.type == type) {
                        outcome = true;
                    } else {
                        return false;
                    }
                }

                return outcome;
            }
            return true;
        },
        [type, region]
    );

    useEffect(() => {
        GetAWSResources(setAWSResources)
    }, []);

    useEffect(() => {
        setupGridData();
    }, [resources]);

    useEffect(() => {
        gridRef?.current?.api?.onFilterChanged();
    }, [type, region]);

    return (
        <Box sx={{ mt: 4 }}>
            {resources
                ? <Box>
                    <Box>
                        <Grid container spacing={2}>
                            <Grid xs={4}>
                                <FilterSelect items={resourceTypes} onChange={onChangeResourceTypes} selected={type} label="Type"></FilterSelect>
                            </Grid>
                            <Grid xs={4}>
                                <FilterSelect items={regions} onChange={onChangeRegions} selected={region} label="Region"></FilterSelect>
                            </Grid>
                            <Grid xs={4}>
                                <FilterSelect items={accounts} onChange={onChangeAccounts} selected={account} label="Account"></FilterSelect>
                            </Grid>
                        </Grid>
                    </Box>

                    <Box className="ag-theme-material" height="80vh" sx={{ mt: 4 }}>
                        <AgGridReact<AWSResource>
                            ref={gridRef}
                            rowData={rows}
                            columnDefs={columns}
                            pagination={true}
                            paginationAutoPageSize={true}
                            animateRows={true}
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
