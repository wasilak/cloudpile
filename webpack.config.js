const Dotenv = require('dotenv-webpack');
const HtmlWebpackPlugin = require('html-webpack-plugin');
const path = require('path');
module.exports = {
    mode: process.env.NODE_ENV,
    entry: './web/src/app/index.tsx',
    module: {
        rules: [
            {
                test: /\.tsx?$/,
                use: 'ts-loader',
                exclude: /node_modules/,
            },
        ],
    },
    resolve: {
        extensions: ['.tsx', '.ts', '.js'],
    },
    output: {
        filename: 'bundle.js',
        path: path.resolve(__dirname, './web/assets'),
    },
    devServer: {
        port: 3000,
        open: true,
        hot: true
    },
    plugins:
        [
            new HtmlWebpackPlugin({
                template: path.resolve(__dirname, './web/src/header_template.html'),
                hash: true, // Cache busting
                filename: path.resolve(__dirname, './web/views/header.html'),
                publicPath: '/public/assets/'
            }),
            new Dotenv()
        ]
};
