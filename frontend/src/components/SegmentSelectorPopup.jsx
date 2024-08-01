import React from "react";

import cookies from 'nookies';

import { Checkbox } from "@/components/ui/checkbox"

export default function SegmentSelectorPopup() {
    // write a popup to select a segment from a list of individual users
    // must be multi-select checkbox group from shadcn
    const [users, setUsers] = useState([]);
    const [selectedUsers, setSelectedUsers] = useState([]);

    const [segments, setSegments] = useState([]);
    const [selectedSegment, setSelectedSegment] = useState([]);

    return (
        <></>
    )
}
