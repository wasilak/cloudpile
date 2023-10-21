import axios from 'axios';
import { AWSResource } from "./models"

export const GetAWSResources = (setAWSResources: any) => {
    axios.get('/api/list')
        .then(response => {
            let resources: AWSResource = response.data;
            setAWSResources(resources);
        })
        .catch(error => {
            console.log(console.error);
        });
};
